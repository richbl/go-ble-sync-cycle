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
	shutdown_manager "github.com/richbl/go-ble-sync-cycle/internal/services"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

// Application constants
const (
	appPrefix  = "----- -----"
	appName    = "BLE Sync Cycle"
	appVersion = "0.7.1"
)

// appControllers holds the application component controllers for managing speed, video playback,
// and BLE communication
type appControllers struct {
	speedController *speed.SpeedController
	videoPlayer     *video.PlaybackController
	bleController   *ble.BLEController
}

func main() {

	// Hello computer...
	waveHello()

	// Load configuration
	cfg := loadConfig("config.toml")

	// Initialize services
	mgr := initializeUtilityServices(cfg)

	// Initialize controllers
	controllers := initializeControllers(cfg)

	// BLE peripheral discovery and CSC scanning
	bleDevice := controllers.bleScanAndConnect(mgr.Context(), mgr)
	controllers.bleGetServicesAndCharacteristics(mgr.Context(), bleDevice, mgr)

	// Start services
	controllers.startServiceRunners(mgr)

	// Wait patiently for shutdown and then wave goodbye
	mgr.Wait()
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

// waveHello outputs a welcome message
func waveHello() {
	log.Println(appPrefix, "Starting", appName, appVersion)
}

// waveGoodbye outputs a goodbye message and exits the program
func waveGoodbye() {
	log.Println(appPrefix, appName, appVersion, "shutdown complete. Goodbye!")
	os.Exit(0)
}

// initializeUtilityServices initializes the core components of the application, including the
// service manager, exit handler, and logger
func initializeUtilityServices(cfg *config.Config) *shutdown_manager.ShutdownManager {

	// Initialize the service manager with a timeout
	mgr := shutdown_manager.NewShutdownManager(30 * time.Second)
	mgr.Start()

	// Initialize the logger
	logger.Initialize(cfg.App.LogLevel)

	// Set the exit handler for fatal log events
	logger.SetExitHandler(func() {
		mgr.Shutdown()
		waveGoodbye()
	})

	return mgr
}

// initializeControllers initializes the application controllers, including the speed controller,
// video player, and BLE controller
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
func logBLESetupError(err error, msg string, mgr *shutdown_manager.ShutdownManager) {

	if err != context.Canceled {
		logger.Fatal(logger.BLE, msg, err.Error())
	}

	// Time to go... so say goodbye
	mgr.HandleExit()
	waveGoodbye()
}

// bleScanAndConnect scans for a BLE peripheral and connects to it
func (controllers *appControllers) bleScanAndConnect(ctx context.Context, mgr *shutdown_manager.ShutdownManager) bluetooth.Device {

	var scanResult bluetooth.ScanResult
	var connectResult bluetooth.Device
	var err error

	if scanResult, err = controllers.bleController.ScanForBLEPeripheral(ctx); err != nil {
		logBLESetupError(err, "failed to scan for BLE peripheral", mgr)
	}

	if connectResult, err = controllers.bleController.ConnectToBLEPeripheral(ctx, scanResult); err != nil {
		logBLESetupError(err, "failed to connect to BLE peripheral", mgr)
	}

	return connectResult
}

// bleGetServicesAndCharacteristics retrieves BLE services and characteristics
func (controllers *appControllers) bleGetServicesAndCharacteristics(ctx context.Context, connectResult bluetooth.Device, mgr *shutdown_manager.ShutdownManager) {

	var serviceResult []bluetooth.DeviceService
	var err error

	if serviceResult, err = controllers.bleController.GetBLEServices(ctx, connectResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE services", mgr)
	}

	if err = controllers.bleController.GetBLECharacteristics(ctx, serviceResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE characteristics", mgr)
	}

}

// startServiceRunners starts the BLE and video service runners and returns a slice of service runners
func (controllers *appControllers) startServiceRunners(mgr *shutdown_manager.ShutdownManager) {

	// Run the BLE service
	mgr.Run("BLE", func(ctx context.Context) error {
		return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController)
	})

	// Run the video service
	mgr.Run("Video", func(ctx context.Context) error {
		return controllers.videoPlayer.Start(ctx, controllers.speedController)
	})
}
