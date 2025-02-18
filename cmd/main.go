package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	flags "github.com/richbl/go-ble-sync-cycle/internal/flags"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	shutdownmanager "github.com/richbl/go-ble-sync-cycle/internal/services"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

// Application constants
const (
	appPrefix  = "----- -----"
	appName    = "BLE Sync Cycle"
	appVersion = "0.10.0"
	configFile = "config.toml"
)

// appControllers holds the application component controllers for managing speed, video playback,
// and BLE communication
type appControllers struct {
	speedController *speed.Controller
	videoPlayer     *video.PlaybackController
	bleController   *ble.Controller
}

// Common errors
var (
	errVideoPlaybackController = errors.New("failed to create video playback controller")
	errBLEController           = errors.New("failed to create BLE controller")
)

func main() {

	// Hello computer...
	waveHello()

	// Parse for command-line flags
	parseCmdLine()

	// Check for help flag
	checkForHelpFlag()

	// Load configuration
	cfg := loadConfig(configFile)

	// Initialize services
	mgr := initializeUtilityServices(cfg)

	// Initialize controllers
	controllers := initializeControllers(cfg)

	// BLE peripheral discovery and CSC scanning
	bleDevice := controllers.bleScanAndConnect(*mgr.Context(), mgr)
	controllers.bleGetServicesAndCharacteristics(*mgr.Context(), bleDevice, mgr)

	// Start services
	controllers.startServiceRunners(mgr)

	// Wait patiently for shutdown and then wave goodbye
	mgr.Wait()
	waveGoodbye()
}

// parseCmdLine parses and validates command-line flags
func parseCmdLine() {

	if err := flags.ParseArgs(); err != nil {
		log.Println(logger.Red+"[FTL] "+logger.Reset+"[APP] failed to parse command-line flags:", err.Error())
		waveGoodbye()
	}

}

// checkForHelpFlag checks for the help flag passed on the command-line
func checkForHelpFlag() {

	clFlags := flags.GetFlags()
	if clFlags.Help {
		flags.ShowHelp()
		waveGoodbye()
	}

}

// loadConfig loads and validates the TOML configuration file
func loadConfig(file string) *config.Config {

	cfg, err := config.Load(file)
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
func initializeUtilityServices(cfg *config.Config) *shutdownmanager.ShutdownManager {

	// Initialize the service manager with a timeout
	mgr := shutdownmanager.NewShutdownManager(30 * time.Second)
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
		return &appControllers{}, logger.VIDEO, fmt.Errorf("%w: %v", errVideoPlaybackController, err)
	}

	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return &appControllers{}, logger.BLE, fmt.Errorf("%w: %v", errBLEController, err)
	}

	return &appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, logger.APP, nil
}

// logBLESetupError displays BLE setup errors and exits the application
func logBLESetupError(err error, msg string, mgr *shutdownmanager.ShutdownManager) {

	if !errors.Is(err, context.Canceled) {
		logger.Fatal(logger.BLE, msg, err.Error())
	}

	// Time to go... so say goodbye
	mgr.HandleExit()
	waveGoodbye()
}

// bleScanAndConnect scans for a BLE peripheral and connects to it
func (controllers *appControllers) bleScanAndConnect(ctx context.Context, mgr *shutdownmanager.ShutdownManager) bluetooth.Device {

	var scanResult bluetooth.ScanResult
	var connectResult bluetooth.Device
	var err error

	if scanResult, err = controllers.bleController.ScanForBLEPeripheral(ctx); err != nil {
		logBLESetupError(err, "failed to scan for BLE peripheral:", mgr)
	}

	if connectResult, err = controllers.bleController.ConnectToBLEPeripheral(ctx, scanResult); err != nil {
		logBLESetupError(err, "failed to connect to BLE peripheral", mgr)
	}

	return connectResult
}

// bleGetServicesAndCharacteristics retrieves BLE services and characteristics
func (controllers *appControllers) bleGetServicesAndCharacteristics(ctx context.Context, connectResult bluetooth.Device, mgr *shutdownmanager.ShutdownManager) {

	var serviceResult []bluetooth.DeviceService
	var err error

	if serviceResult, err = controllers.bleController.GetBLEServices(ctx, connectResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE services:", mgr)
	}

	if err = controllers.bleController.GetBLECharacteristics(ctx, serviceResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE characteristics", mgr)
	}

}

// startServiceRunners starts the BLE and video service runners and returns a slice of service runners
func (controllers *appControllers) startServiceRunners(mgr *shutdownmanager.ShutdownManager) {

	// Run the BLE service
	mgr.Run(func(ctx context.Context) error {

		if err := controllers.bleController.GetBLEUpdates(ctx, controllers.speedController); err != nil {
			logger.Info(logger.BLE, err.Error())
			return err
		}

		return nil
	})

	// Run the video service
	mgr.Run(func(ctx context.Context) error {

		if err := controllers.videoPlayer.Start(ctx, controllers.speedController); err != nil {
			logger.Info(logger.VIDEO, err.Error())
			return err
		}

		return nil
	})

}
