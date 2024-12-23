package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

// appControllers holds the main application controllers
type appControllers struct {
	speedController *speed.SpeedController
	videoPlayer     *video.PlaybackController
	bleController   *ble.BLEController
}

// shutdownHandler encapsulates shutdown coordination
type shutdownHandler struct {
	done           chan struct{}
	componentsDown chan struct{}
	cleanupOnce    sync.Once
	wg             *sync.WaitGroup
	rootCancel     context.CancelFunc
	resetTerminal  func()
}

// componentErr holds the error type and component type used for logging
type componentErr struct {
	componentType logger.ComponentType
	err           error
}

const (
	appPrefix  = "----- -----"
	appName    = "BLE Sync Cycle"
	appVersion = "0.6.2"
)

func main() {
	// Hello computer!
	log.Println(appPrefix, "Starting", appName, appVersion)

	// Load configuration from TOML
	cfg := loadConfig("config.toml")

	// Initialize logger package with display level from TOML configuration
	logger.Initialize(cfg.App.LogLevel)

	// Initialize shutdown services: signal handling, context cancellation and terminal config
	var wg sync.WaitGroup
	rootCtx, rootCancel := context.WithCancel(context.Background())
	sh := &shutdownHandler{
		done:           make(chan struct{}),
		componentsDown: make(chan struct{}),
		wg:             &wg,
		rootCancel:     rootCancel,
		resetTerminal:  configTerminal(),
	}
	defer rootCancel()

	// Set up shutdown handlers
	setupSignalHandling(sh)
	logger.SetExitHandler(sh.cleanup)

	// Create component controllers
	controllers, componentType, err := setupAppControllers(*cfg)
	if err != nil {
		logger.Fatal(componentType, "failed to create controllers: "+err.Error())
		<-sh.done
		waveGoodbye()
	}

	// Start components
	if componentType, err := startAppControllers(rootCtx, controllers, sh.wg); err != nil {
		logger.Fatal(componentType, err.Error())
		<-sh.done
		waveGoodbye()
	}

	<-sh.done
}

// loadConfig loads the TOML configuration file
func loadConfig(file string) *config.Config {
	cfg, err := config.LoadFile(file)
	if err != nil {
		log.Println(logger.Red + "[FTL]" + logger.Reset + " [APP] failed to load TOML configuration: " + err.Error())
		waveGoodbye()
	}

	return cfg
}

// cleanup handles graceful shutdown of all components
func (sh *shutdownHandler) cleanup() {

	sh.cleanupOnce.Do(func() {
		// Signal components to shut down and wait for them to finish
		sh.rootCancel()
		sh.wg.Wait()
		close(sh.componentsDown)

		// Perform final cleanup
		sh.resetTerminal()
		close(sh.done)
		waveGoodbye()
	})

}

// waveGoodbye outputs a goodbye message and exits the application
func waveGoodbye() {
	log.Println(appPrefix, appName, appVersion, "shutdown complete. Goodbye!")
	os.Exit(0)
}

// setupSignalHandling configures OS signal handling
func setupSignalHandling(sh *shutdownHandler) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		sh.cleanup()
	}()

}

// configTerminal handles terminal char echo to prevent display of break (^C) character
func configTerminal() func() {
	// Disable control character echo using stty
	rawMode := exec.Command("stty", "-echo")
	rawMode.Stdin = os.Stdin

	if err := rawMode.Run(); err != nil {
		logger.Fatal(logger.APP, "failed to configure terminal: "+err.Error())
		waveGoodbye()
	}

	// Return cleanup function
	return func() {
		cooked := exec.Command("stty", "echo")
		cooked.Stdin = os.Stdin

		if err := cooked.Run(); err != nil {
			logger.Fatal(logger.APP, "failed to restore terminal: "+err.Error())
			waveGoodbye()
		}

	}

}

// setupAppControllers creates and initializes the application controllers
func setupAppControllers(cfg config.Config) (appControllers, logger.ComponentType, error) {
	// Create speed and video controllers
	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return appControllers{}, logger.VIDEO, errors.New("failed to create video player: " + err.Error())
	}

	// Create BLE controller
	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return appControllers{}, logger.BLE, errors.New("failed to create BLE controller: " + err.Error())
	}

	return appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, logger.APP, nil
}

// startAppControllers is responsible for starting and managing the component controllers
func startAppControllers(ctx context.Context, controllers appControllers, wg *sync.WaitGroup) (logger.ComponentType, error) {
	// Create shutdown signal context.
	ctxWithCancel, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Scan to find the speed characteristic
	bleSpeedCharacter, err := runBLEScan(ctxWithCancel, controllers)
	if err != nil {

		// Check if the context was cancelled (user pressed Ctrl+C)
		if errors.Is(err, context.Canceled) {
			return logger.APP, nil
		}

		return logger.BLE, errors.New("BLE peripheral scan failed: " + err.Error())
	}

	errChan := make(chan componentErr, 2) // Buffered channel for component errors
	wg.Add(2)

	// Start BLE monitoring and video playback
	startBLEMonitoring(ctxWithCancel, controllers, wg, bleSpeedCharacter, errChan)
	startVideoPlaying(ctxWithCancel, controllers, wg, errChan)

	// Wait for component results or cancellation
	for i := 0; i < 2; i++ {

		select {
		case compErr := <-errChan:

			if compErr.err != nil {
				return compErr.componentType, compErr.err
			}

		// Context cancelled, no error
		case <-ctxWithCancel.Done():
			return logger.APP, nil
		}

	}

	return logger.APP, nil
}

// runBLEScan scans for the BLE speed characteristic.
func runBLEScan(ctx context.Context, controllers appControllers) (*bluetooth.DeviceCharacteristic, error) {
	results := make(chan *bluetooth.DeviceCharacteristic, 1)
	errChan := make(chan error, 1)

	go func() {
		characteristic, err := controllers.bleController.GetBLECharacteristic(ctx, controllers.speedController)
		if err != nil {
			errChan <- err
			return
		}

		results <- characteristic
	}()

	select {
	case <-ctx.Done():
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE characteristic scan...")
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case characteristic := <-results:
		return characteristic, nil
	}
}

// startBLEMonitoring starts the BLE monitoring goroutine
func startBLEMonitoring(ctx context.Context, controllers appControllers, wg *sync.WaitGroup, bleSpeedCharacter *bluetooth.DeviceCharacteristic, errChan chan<- componentErr) {
	go func() {
		defer wg.Done()

		if err := monitorBLESpeed(ctx, controllers, bleSpeedCharacter); err != nil {

			// Only send error if context was not cancelled
			if !errors.Is(err, context.Canceled) {
				errChan <- componentErr{componentType: logger.BLE, err: err}
			}

			return
		}

		errChan <- componentErr{componentType: logger.BLE, err: nil}
	}()
}

// startVideoPlaying starts the video playing goroutine.
func startVideoPlaying(ctx context.Context, controllers appControllers, wg *sync.WaitGroup, errChan chan<- componentErr) {
	go func() {
		defer wg.Done()

		if err := playVideo(ctx, controllers); err != nil {

			// Only send error if context was not cancelled
			if !errors.Is(err, context.Canceled) {
				errChan <- componentErr{componentType: logger.VIDEO, err: err}
			}

			return
		}

		errChan <- componentErr{componentType: logger.VIDEO, err: nil}
	}()
}

// monitorBLESpeed monitors the BLE speed characteristic
func monitorBLESpeed(ctx context.Context, controllers appControllers, bleSpeedCharacter *bluetooth.DeviceCharacteristic) error {
	return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController, bleSpeedCharacter)
}

// playVideo starts the video player.
func playVideo(ctx context.Context, controllers appControllers) error {
	return controllers.videoPlayer.Start(ctx, controllers.speedController)
}
