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
	"tinygo.org/x/bluetooth"
)

// Application constants
const (
	appPrefix       = "----- -----"
	appName         = "BLE Sync Cycle"
	appVersion      = "0.7.0"
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
	controllers := initializeControllers(cfg)

	// BLE peripheral discovery and CSC scanning
	bleDevice := controllers.bleScanAndConnect(sm.shutdownCtx.ctx, exitHandler)
	controllers.bleGetServicesAndCharacteristics(sm.shutdownCtx.ctx, bleDevice, exitHandler)

	// Start and monitor services for BLE and video components
	monitorServiceRunners(controllers.startServiceRunners(sm))

	// Wait for final shutdown sequences to complete and wave goodbye!
	sm.Wait()
	waveGoodbye()
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
func initializeControllers(cfg *config.Config) *appControllers {

	controllers, componentType, err := setupAppControllers(*cfg)
	if err != nil {
		logger.Fatal(componentType, "failed to create controllers:", err.Error())
	}

	return controllers
}

// setupAppControllers creates and initializes all application controllers
func setupAppControllers(cfg config.Config) (*appControllers, logger.ComponentType, error) {

	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return &appControllers{}, logger.VIDEO, fmt.Errorf("failed to create video playback controller: %v", err)
	}

	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return &appControllers{}, logger.BLE, fmt.Errorf("failed to create BLE controller: %v", err)
	}

	return &appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, logger.APP, nil
}

// logBLESetupError displays BLE setup errors and exits the application
func logBLESetupError(err error, msg string, exitHandler *ExitHandler) {

	if err != context.Canceled {
		logger.Fatal(logger.BLE, msg, err.Error())
	}

	exitHandler.HandleExit()
}

// bleScanAndConnect scans for a BLE peripheral and connects to it
func (controllers *appControllers) bleScanAndConnect(ctx context.Context, exitHandler *ExitHandler) bluetooth.Device {

	var scanResult bluetooth.ScanResult
	var connectResult bluetooth.Device
	var err error

	if scanResult, err = controllers.bleController.ScanForBLEPeripheral(ctx); err != nil {
		logBLESetupError(err, "failed to scan for BLE peripheral", exitHandler)
	}

	if connectResult, err = controllers.bleController.ConnectToBLEPeripheral(ctx, scanResult); err != nil {
		logBLESetupError(err, "failed to connect to BLE peripheral", exitHandler)
	}

	return connectResult
}

// bleGetServicesAndCharacteristics retrieves BLE services and characteristics
func (controllers *appControllers) bleGetServicesAndCharacteristics(ctx context.Context, connectResult bluetooth.Device, exitHandler *ExitHandler) {

	var serviceResult []bluetooth.DeviceService
	var err error

	if serviceResult, err = controllers.bleController.GetBLEServices(ctx, connectResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE services", exitHandler)
	}

	if err = controllers.bleController.GetBLECharacteristics(ctx, serviceResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE characteristics", exitHandler)
	}

}

// startServiceRunners starts the BLE and video service runners and returns a slice of service runners
func (controllers *appControllers) startServiceRunners(sm *ShutdownManager) []*ServiceRunner {

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

// monitorServiceRunners monitors the services and logs any errors encountered
func monitorServiceRunners(runners []*ServiceRunner) {

	for _, runner := range runners {

		if err := runner.Error(); err != nil {
			logger.Fatal(logger.APP, "service error:", err.Error())
			return
		}

	}
}
