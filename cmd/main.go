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
	restoreTerm    func()
}

func main() {
	log.Println("----- ----- Starting BLE Sync Cycle 0.6.2")

	// Load configuration
	cfg, err := config.LoadFile("config.toml")
	if err != nil {
		log.Println(logger.Red + "[FTL]" + logger.Reset + " [APP] failed to load TOML configuration: " + err.Error())
		log.Println("----- ----- BLE Sync Cycle 0.6.2 shutdown complete. Goodbye!")
		os.Exit(0)
	}

	logger.Initialize(cfg.App.LogLevel)

	// Initialize shutdown coordination
	var wg sync.WaitGroup
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	sh := &shutdownHandler{
		done:           make(chan struct{}),
		componentsDown: make(chan struct{}),
		wg:             &wg,
		rootCancel:     rootCancel,
		restoreTerm:    configureTerminal(),
	}

	// Set up shutdown handlers
	setupSignalHandling(sh)
	logger.SetExitHandler(sh.cleanup)

	// Create component controllers
	controllers, componentType, err := setupAppControllers(*cfg)
	if err != nil {
		logger.Fatal(componentType, "failed to create controllers: "+err.Error())
		<-sh.done
		os.Exit(0)
	}

	// Start components
	if componentType, err := startAppControllers(rootCtx, controllers, sh.wg); err != nil {
		logger.Error(componentType, err.Error())
		sh.cleanup()
		<-sh.done
		os.Exit(0)
	}

	<-sh.done
}

// cleanup handles graceful shutdown of all components
func (sh *shutdownHandler) cleanup() {

	sh.cleanupOnce.Do(func() {

		// Signal components to shut down and wait for them to finish
		sh.rootCancel()
		sh.wg.Wait()
		close(sh.componentsDown)

		// Perform final cleanup
		sh.restoreTerm()

		log.Println("----- ----- BLE Sync Cycle 0.6.2 shutdown complete. Goodbye!")
		close(sh.done)
	})
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

// configureTerminal handles terminal char echo to prevent display of break (^C) character
func configureTerminal() func() {
	// Disable control character echo using stty
	rawMode := exec.Command("stty", "-echo")
	rawMode.Stdin = os.Stdin
	_ = rawMode.Run()

	// Return cleanup function
	return func() {
		cooked := exec.Command("stty", "echo")
		cooked.Stdin = os.Stdin
		_ = cooked.Run()
	}
}

// setupAppControllers creates and initializes the application controllers
func setupAppControllers(cfg config.Config) (appControllers, logger.ComponentType, error) {
	// Create speed  and video controllers
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
	// componentErr holds the error type and component type used for logging
	type componentErr struct {
		componentType logger.ComponentType
		err           error
	}

	// Create shutdown signal
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Scan for BLE peripheral of interest
	bleSpeedCharacter, err := scanForBLESpeedCharacteristic(ctx, controllers)
	if err != nil {

		// Check if the context was cancelled (user pressed Ctrl+C)
		if ctx.Err() == context.Canceled {
			return logger.APP, nil
		}

		return logger.BLE, errors.New("BLE peripheral scan failed: " + err.Error())
	}

	// Start component controllers concurrently
	errs := make(chan componentErr, 1)

	// Add the goroutines to WaitGroup before starting them
	wg.Add(2)
	// Monitor BLE speed (goroutine)
	go func() {
		defer wg.Done()

		if err := monitorBLESpeed(ctx, controllers, bleSpeedCharacter); err != nil {

			// Check if the context was cancelled (user pressed Ctrl+C)
			if ctx.Err() == context.Canceled {
				errs <- componentErr{logger.BLE, nil}
				return
			}

			errs <- componentErr{logger.BLE, err}
			return
		}

		errs <- componentErr{logger.BLE, nil}
	}()

	// Play video (goroutine)
	go func() {
		defer wg.Done()

		if err := playVideo(ctx, controllers); err != nil {

			// Check if the context was cancelled (user pressed Ctrl+C)
			if ctx.Err() == context.Canceled {
				errs <- componentErr{logger.VIDEO, nil}
				return
			}

			errs <- componentErr{logger.VIDEO, err}
			return
		}

		errs <- componentErr{logger.VIDEO, nil}
	}()

	// Wait for both component results
	for i := 0; i < 2; i++ {
		compErr := <-errs

		if compErr.err != nil {
			return compErr.componentType, compErr.err
		}

	}

	return logger.APP, nil
}

// scanForBLESpeedCharacteristic scans for the BLE CSC speed characteristic
func scanForBLESpeedCharacteristic(ctx context.Context, controllers appControllers) (*bluetooth.DeviceCharacteristic, error) {
	// create a channel to receive the characteristic
	results := make(chan *bluetooth.DeviceCharacteristic, 1)
	errChan := make(chan error, 1)

	// Scan for the BLE CSC speed characteristic
	go func() {
		characteristic, err := controllers.bleController.GetBLECharacteristic(ctx, controllers.speedController)

		if err != nil {
			errChan <- err
			return
		}

		// Return the characteristic
		results <- characteristic
	}()

	// Wait for the characteristic or an error
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

// monitorBLESpeed monitors the BLE speed characteristic
func monitorBLESpeed(ctx context.Context, controllers appControllers, bleSpeedCharacter *bluetooth.DeviceCharacteristic) error {
	return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController, bleSpeedCharacter)
}

// playVideo starts the video player
func playVideo(ctx context.Context, controllers appControllers) error {
	return controllers.videoPlayer.Start(ctx, controllers.speedController)
}
