package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"
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

	// Hello world!
	log.Println(appPrefix, "Starting", appName, appVersion)

	// Load configuration
	cfg := loadConfig("config.toml")

	// Initialize utility services
	sm, exitHandler := initializeUtilityServices(cfg)

	// Initialize application controllers
	controllers := initializeControllers(cfg, exitHandler)

	// Scan for BLE device
	bleDeviceDiscovery(sm.shutdownCtx.ctx, controllers, exitHandler)

	// Start and monitor services for BLE and video components
	monitorServiceRunners(startServiceRunners(sm, controllers))

	// Wait for final shutdown sequences to complete and wave goodbye!
	sm.Wait()
	waveGoodbye()
}

// monitorServiceRunners monitors the services and logs any errors encountered
func monitorServiceRunners(runners []*ServiceRunner) {

	for _, runner := range runners {

		if err := runner.Error(); err != nil {
			logger.Fatal(logger.APP, "service error:", err.Error())
			return
		}

	}
}

// startServiceRunners starts the BLE and video service runners and returns a slice of service runners
func startServiceRunners(sm *ShutdownManager, controllers appControllers) []*ServiceRunner {

	// Create and run the BLE service runner
	bleRunner := NewServiceRunner(sm, "BLE")
	bleRunner.Run(func(ctx context.Context) error {
		return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController)
	})

	// Create and run the video service runner
	videoRunner := NewServiceRunner(sm, "Video")
	videoRunner.Run(func(ctx context.Context) error {
		return controllers.videoPlayer.Start(ctx, controllers.speedController)
	})

	return []*ServiceRunner{bleRunner, videoRunner}
}

// bleDeviceDiscovery scans for the BLE device and CSC speed characteristic
func bleDeviceDiscovery(ctx context.Context, controllers appControllers, exitHandler *ExitHandler) {

	err := scanForBLECharacteristic(ctx, controllers)
	if err != nil {

		if err != context.Canceled {
			logger.Fatal(logger.BLE, "failed to scan for BLE characteristic:", err.Error())
			return
		}

		exitHandler.HandleExit()
	}
}

// initializeUtilityServices initializes the core components of the application, including the shutdown manager,
// exit handler, and logger
func initializeUtilityServices(cfg *config.Config) (*ShutdownManager, *ExitHandler) {

	// Initialize the shutdown manager and exit handler
	sm := NewShutdownManager(shutdownTimeout)
	exitHandler := NewExitHandler(sm)
	sm.Start()

	// Initialize the logger
	logger.Initialize(cfg.App.LogLevel)

	// Set the exit handler for the shutdown manager
	logger.SetExitHandler(func() {
		sm.initiateShutdown()
		exitHandler.HandleExit()
	})

	return sm, exitHandler
}

// initializeControllers initializes the application controllers, including the speed controller,
// video player, and BLE controller. It returns the initialized controllers
func initializeControllers(cfg *config.Config, exitHandler *ExitHandler) appControllers {

	controllers, componentType, err := setupAppControllers(*cfg)

	if err != nil {
		logger.Fatal(componentType, "failed to create controllers:", err.Error())
		exitHandler.HandleExit()
	}

	return controllers
}

// setupAppControllers creates and initializes all application controllers
func setupAppControllers(cfg config.Config) (appControllers, logger.ComponentType, error) {

	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return appControllers{}, logger.VIDEO, fmt.Errorf("failed to create video playback controller: %v", err)
	}

	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return appControllers{}, logger.BLE, fmt.Errorf("failed to create BLE controller: %v", err)
	}

	return appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, logger.APP, nil
}

// scanForBLECharacteristic handles the initial BLE device discovery and characteristic scanning
func scanForBLECharacteristic(ctx context.Context, controllers appControllers) error {

	// Create a channel to receive errors from the scan goroutine
	errChan := make(chan error, 1)

	// BLE peripheral scan and connect
	go func() {
		defer close(errChan)
		scanResult, err := controllers.bleController.ScanForBLEPeripheral(ctx)
		if err != nil {
			errChan <- err
			return
		}

		connectResult, err := controllers.bleController.ConnectToBLEPeripheral(scanResult)
		if err != nil {
			errChan <- err
			return
		}

		// Get the BLE characteristic from the connected device
		err = controllers.bleController.GetBLECharacteristic(connectResult)
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		fmt.Print("\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE discovery...")
		return ctx.Err()
	case result := <-errChan:
		return result
	}
}

// loadConfig loads and validates the TOML configuration file
func loadConfig(file string) *config.Config {

	cfg, err := config.LoadFile(file)
	if err != nil {
		log.Println(logger.Red+"[FTL] "+logger.Reset+"[APP] failed to load TOML configuration:", err.Error())
		waveGoodbye()
	}

	return cfg
}

// waveGoodbye outputs a goodbye message and exits the program
func waveGoodbye() {
	log.Println(appPrefix, appName, appVersion, "shutdown complete. Goodbye!")
	os.Exit(0)
}
