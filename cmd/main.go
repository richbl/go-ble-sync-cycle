package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"time"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

// Application constants
const (
	appPrefix       = "----- -----"
	appName         = "BLE Sync Cycle"
	appVersion      = "0.6.2"
	shutdownTimeout = 30 * time.Second
)

// appControllers holds the application component controllers for managing speed, video playback,
// and BLE communication
type appControllers struct {
	speedController *speed.SpeedController
	videoPlayer     *video.PlaybackController
	bleController   *ble.BLEController
}

func main() {
	log.Println(appPrefix, "Starting", appName, appVersion)

	cfg := loadConfig("config.toml")

	// Initialize the shutdown manager and exit handler
	sm := NewShutdownManager(shutdownTimeout)
	exitHandler := NewExitHandler(sm)

	// Add configureTerminal cleanup function to reset terminal settings on exit
	sm.AddCleanupFn(configureTerminal())
	sm.Start()

	// Initialize the logger with the configured log level and exit handler
	logger.Initialize(cfg.App.LogLevel)
	logger.SetExitHandler(func() {
		sm.initiateShutdown()
		exitHandler.HandleExit()
	})

	// Initialize the application controllers
	controllers, componentType, err := setupAppControllers(*cfg)
	if err != nil {
		logger.Fatal(componentType, "failed to create controllers: "+err.Error())
		return
	}

	// Scan for the BLE characteristic and handle context cancellation
	bleChar, err := scanForBLECharacteristic(sm.Context(), controllers)
	if err != nil {

		if err != context.Canceled {
			logger.Fatal(logger.BLE, "failed to scan for BLE characteristic: "+err.Error())
			return
		}

		exitHandler.HandleExit()
		return
	}

	// Create and run the BLE service runner
	bleRunner := NewServiceRunner(sm, "BLE")
	bleRunner.Run(func(ctx context.Context) error {
		return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController, bleChar)
	})

	// Create and run the video service runner
	videoRunner := NewServiceRunner(sm, "Video")
	videoRunner.Run(func(ctx context.Context) error {
		return controllers.videoPlayer.Start(ctx, controllers.speedController)
	})

	// Wait for services to complete and check for errors
	for _, runner := range []*ServiceRunner{bleRunner, videoRunner} {
		if err := runner.Error(); err != nil {
			logger.Fatal(logger.APP, "service error: "+err.Error())
			return
		}
	}

	// Wait for final shutdown sequences to complete and wave goodbye!
	sm.Wait()
	waveGoodbye()
}

// setupAppControllers creates and initializes all application controllers
func setupAppControllers(cfg config.Config) (appControllers, logger.ComponentType, error) {

	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return appControllers{}, logger.VIDEO, errors.New("failed to create video player: " + err.Error())
	}

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

// scanForBLECharacteristic handles the initial BLE device discovery and characteristic scanning
// using a context for cancellation and returns the discovered characteristic or an error
func scanForBLECharacteristic(ctx context.Context, controllers appControllers) (*bluetooth.DeviceCharacteristic, error) {

	// Create a channel to receive the result of the BLE characteristic scan
	resultsChan := make(chan struct {
		char *bluetooth.DeviceCharacteristic
		err  error
	}, 1)

	go func() {
		defer close(resultsChan)
		char, err := controllers.bleController.GetBLECharacteristic(ctx, controllers.speedController)
		resultsChan <- struct {
			char *bluetooth.DeviceCharacteristic
			err  error
		}{char, err}
	}()

	select {
	case <-ctx.Done():
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE discovery...")
		return nil, ctx.Err()
	case result := <-resultsChan:
		return result.char, result.err
	}
}

// loadConfig loads and validates the TOML configuration file
func loadConfig(file string) *config.Config {

	cfg, err := config.LoadFile(file)
	if err != nil {
		log.Println(logger.Red + "[FTL]" + logger.Reset + " [APP] failed to load TOML configuration: " + err.Error())
		waveGoodbye()
	}

	return cfg
}

// configureTerminal handles terminal character echo settings, returning a cleanup function
// to restore original terminal settings
func configureTerminal() func() {

	rawMode := exec.Command("stty", "-echo")
	rawMode.Stdin = os.Stdin
	_ = rawMode.Run()

	return func() {
		cooked := exec.Command("stty", "echo")
		cooked.Stdin = os.Stdin
		_ = cooked.Run()
	}
}

// waveGoodbye outputs a goodbye message and exits the program
func waveGoodbye() {
	log.Println(appPrefix, appName, appVersion, "shutdown complete. Goodbye!")
	os.Exit(0)
}
