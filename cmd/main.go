package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	flags "github.com/richbl/go-ble-sync-cycle/internal/flags"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	services "github.com/richbl/go-ble-sync-cycle/internal/services"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video"
	"github.com/richbl/go-ble-sync-cycle/ui"

	"tinygo.org/x/bluetooth"
)

// Application constants
const (
	appName    = "BLE Sync Cycle"
	appVersion = "0.13.0"
	configFile = "config.toml"
	errFormat  = "%v: %w"
)

// appControllers holds the application component controllers for managing speed, video playback,
// and BLE communication
type appControllers struct {
	speedController *speed.Controller
	videoPlayer     *video.PlaybackController
	bleController   *ble.Controller
}

// Error definitions
var (
	errVideoPlaybackController = errors.New("failed to create video playback controller")
	errBLEController           = errors.New("failed to create BLE controller")
)

func main() {

	// Parse for command-line flags
	parseCmdLine()

	// Initialize the default logger until config is loaded
	logger.Initialize("debug")

	// Hello computer...
	waveHello()

	// Check for help flag
	checkForHelpFlag()

	// Check for application mode (GUI or no-GUI)
	if !checkForNoGUIFlag() {
		logger.Info(logger.APP, "running in GUI mode: GUI features enabled")
		ui.StartGUI()

		return // Exit CLI logic to allow GUI to take over
	}

	// Load configuration from TOML file
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
		logger.Fatal(logger.APP, "failed to parse command-line flags:", err.Error())
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

// checkForNoGUIFlag checks for the no-gui flag passed on the command-line
func checkForNoGUIFlag() bool {
	clFlags := flags.GetFlags()
	return clFlags.NoGUI
}

// loadConfig loads and validates the TOML configuration file
func loadConfig(file string) *config.Config {

	cfg, err := config.Load(file)
	if err != nil {
		logger.Fatal(logger.APP, "failed to load TOML configuration:", err.Error())
	}

	return cfg
}

// waveHello outputs a welcome message
func waveHello() {
	logger.Info(logger.APP, "Starting", appName, appVersion)
}

// waveGoodbye outputs a goodbye message and exits the program
func waveGoodbye() {
	logger.Info(logger.APP, appName, appVersion, "shutdown complete. Goodbye")
	os.Exit(0)
}

// initializeUtilityServices initializes the core components of the application, including the
// service manager, exit handler, and logger
func initializeUtilityServices(cfg *config.Config) *services.ShutdownManager {

	// Initialize the service manager with a timeout
	mgr := services.NewShutdownManager(30 * time.Second)
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
		return &appControllers{}, logger.VIDEO, fmt.Errorf(errFormat, errVideoPlaybackController, err)
	}

	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return &appControllers{}, logger.BLE, fmt.Errorf(errFormat, errBLEController, err)
	}

	return &appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, logger.APP, nil
}

// logBLESetupError displays BLE setup errors and exits the application
func logBLESetupError(err error, msg string, mgr *services.ShutdownManager) {

	if !errors.Is(err, context.Canceled) {
		logger.Fatal(logger.BLE, msg+":", err.Error())
	}

	// Time to go... so say goodbye
	mgr.HandleExit()
	waveGoodbye()
}

// bleScanAndConnect scans for a BLE peripheral and connects to it
func (controllers *appControllers) bleScanAndConnect(ctx context.Context, mgr *services.ShutdownManager) bluetooth.Device {

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
func (controllers *appControllers) bleGetServicesAndCharacteristics(ctx context.Context, connectResult bluetooth.Device, mgr *services.ShutdownManager) {

	var serviceResult []ble.CharacteristicDiscoverer
	var err error

	// Get the battery service
	if serviceResult, err = controllers.bleController.GetBatteryService(ctx, &connectResult); err != nil {
		logBLESetupError(err, "failed to acquire battery service", mgr)
	}

	// Get the battery level
	if err = controllers.bleController.GetBatteryLevel(ctx, serviceResult); err != nil {
		logBLESetupError(err, "failed to acquire battery level", mgr)
	}

	// Get the CSC services
	if serviceResult, err = controllers.bleController.GetCSCServices(ctx, &connectResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE services", mgr)
	}

	// Get the CSC characteristics
	if err = controllers.bleController.GetCSCCharacteristics(ctx, serviceResult); err != nil {
		logBLESetupError(err, "failed to acquire BLE characteristics", mgr)
	}

}

// startServiceRunners starts the BLE and video service runners and returns a slice of service runners
func (controllers *appControllers) startServiceRunners(mgr *services.ShutdownManager) {

	// Run the BLE service
	mgr.Run(func(ctx context.Context) error {

		if err := controllers.bleController.GetBLEUpdates(ctx, controllers.speedController); err != nil {
			logger.Error(logger.BLE, err.Error())
			return err
		}

		return nil
	})

	// Run the video service
	mgr.Run(func(ctx context.Context) error {

		if err := controllers.videoPlayer.Start(ctx, controllers.speedController); err != nil {
			logger.Error(logger.VIDEO, err.Error())
			return err
		}

		return nil
	})

}
