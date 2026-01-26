package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/ble"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
	"github.com/richbl/go-ble-sync-cycle/internal/speed"
	"github.com/richbl/go-ble-sync-cycle/internal/video"
	"tinygo.org/x/bluetooth"
)

// Error definitions
var (
	errNoActiveConfig  = errors.New("cannot initialize controllers: no active configuration")
	errNoActiveSession = errors.New("no active session to stop")
)

// controllers holds the application component controllers
type controllers struct {
	speedController *speed.Controller
	videoPlayer     *video.PlaybackController
	bleController   *ble.Controller
	bleDevice       bluetooth.Device
}

// StartSession initializes controllers and starts BLE and video services
func (m *StateManager) StartSession() error {

	// Validate preconditions and flip PendingStart/state to Connecting atomically
	if err := m.prepareStart(); err != nil {
		return err
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating ShutdownManager")

	shutdownMgr := services.NewShutdownManager(30 * time.Second)
	shutdownMgr.Start()

	// store the shutdown manager reference
	m.storeShutdownMgr(shutdownMgr)

	// store the context for the now-sentient shutdown manager
	ctx := *shutdownMgr.Context()

	logger.Debug(ctx, logger.APP, "initializing controllers...")

	controllers, err := m.initializeControllers()
	if err != nil {
		logger.Error(ctx, logger.APP, fmt.Sprintf("controllers init failed: %v", err))
		m.cleanupStartFailure(shutdownMgr)

		return fmt.Errorf("failed to initialize controllers: %w", err)
	}

	logger.Debug(ctx, logger.APP, "controllers initialized OK")
	logger.Debug(ctx, logger.APP, "connecting BLE...")

	bleDevice, err := m.connectBLE(controllers, shutdownMgr)
	if err != nil {
		logger.Error(ctx, logger.APP, fmt.Sprintf("BLE connect failed: %v", err))
		m.cleanupStartFailure(shutdownMgr)

		return fmt.Errorf("BLE connection failed: %w", err)
	}

	controllers.bleDevice = bleDevice

	logger.Debug(ctx, logger.APP, "BLE connected OK")

	// Finalize successful start
	m.mu.Lock()
	m.controllers = controllers
	m.PendingStart = false
	m.state = StateRunning
	logger.Debug(ctx, logger.APP, "set state=Running, PendingStart=false")
	m.mu.Unlock()

	logger.Debug(ctx, logger.APP, "starting services...")

	m.startServices(controllers, shutdownMgr)

	logger.Debug(ctx, logger.APP, "services started")

	return nil
}

// StopSession stops all services and cleans up controllers
func (m *StateManager) StopSession() error {

	m.mu.Lock()

	shutdownMgr := m.shutdownMgr
	wasPending := m.PendingStart

	// Log what we're destroying
	m.logControllersRelease(shutdownMgr)

	m.state = StateLoaded
	m.PendingStart = false
	m.mu.Unlock()

	if shutdownMgr == nil && !wasPending {
		return errNoActiveSession
	}

	ctx := logger.BackgroundCtx
	if shutdownMgr != nil {
		ctx = *shutdownMgr.Context()
	}

	if wasPending {
		logger.Info(ctx, logger.APP, "stop requested, canceling pending session setup...")
	} else {
		logger.Info(ctx, logger.APP, "stop requested, canceling active session...")
	}

	fmt.Fprint(os.Stdout, "\r") // Clear the ^C character from the terminal line

	// Stop the shutdown manager
	if shutdownMgr != nil {
		shutdownMgr.Shutdown()
	}

	m.clearResources()
	m.stopBLEScan(ctx)

	if wasPending {
		logger.Info(ctx, logger.APP, "stopped pending session startup")
	} else {
		logger.Info(ctx, logger.APP, "session stopped")
	}

	return nil
}

// logControllersRelease logs the release of controller objects
func (m *StateManager) logControllersRelease(shutdownMgr *services.ShutdownManager) {

	if m.controllers == nil || shutdownMgr == nil {
		return
	}

	ctx := *shutdownMgr.Context()

	if m.controllers.bleController != nil {
		logger.Info(ctx, logger.BLE, fmt.Sprintf("releasing BLE controller object (id:%04d)", m.controllers.bleController.InstanceID))
	}
	if m.controllers.speedController != nil {
		logger.Info(ctx, logger.SPEED, fmt.Sprintf("releasing speed controller object (id:%04d)", m.controllers.speedController.InstanceID))
	}
	if m.controllers.videoPlayer != nil {
		logger.Info(ctx, logger.VIDEO, fmt.Sprintf("releasing video controller object (id:%04d)", m.controllers.videoPlayer.InstanceID))
	}

}

// clearResources clears the session resources
func (m *StateManager) clearResources() {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.controllers = nil
	m.shutdownMgr = nil
	m.activeConfig = nil

	logger.Debug(logger.BackgroundCtx, logger.APP, "controllers and shutdown manager objects released")

}

// stopBLEScan stops the BLE scan
func (m *StateManager) stopBLEScan(ctx context.Context) {

	// Stop any ongoing scan under mutex
	ble.AdapterMu.Lock()
	defer ble.AdapterMu.Unlock()

	if err := bluetooth.DefaultAdapter.StopScan(); err != nil {
		logger.Warn(ctx, logger.BLE, fmt.Sprintf("failed to stop current BLE scan: %v", err))
	} else {
		logger.Info(ctx, logger.BLE, "BLE scan stopped")
	}

}

// BatteryLevel returns the current battery level from the BLE controller
func (m *StateManager) BatteryLevel() byte {

	defer m.readLock()()

	if m.controllers != nil && m.controllers.bleController != nil {
		return m.controllers.bleController.BatteryLevelLast()
	}

	return 0 // Unknown (0%)
}

// CurrentSpeed returns the current smoothed speed from the speed controller
func (m *StateManager) CurrentSpeed() (float64, string) {

	defer m.readLock()()

	// Use ActiveConfig here to ensure we return the units of the active running session
	cfg := m.activeConfig
	if cfg == nil {
		cfg = m.editConfig
	}

	// Check for nil controllers (session stopped or not started)
	if m.controllers == nil || m.controllers.speedController == nil || cfg == nil {
		return 0.0, ""
	}

	return m.controllers.speedController.SmoothedSpeed(), cfg.Speed.SpeedUnits
}

// VideoTimeRemaining returns the formatted time remaining string (HH:MM:SS)
func (m *StateManager) VideoTimeRemaining() string {

	defer m.readLock()()

	noTime := "--:--:--"

	if m.controllers == nil || m.controllers.videoPlayer == nil {
		return noTime
	}

	timeStr, err := m.controllers.videoPlayer.TimeRemaining()
	if err != nil {
		return noTime
	}

	return timeStr
}

// VideoPlaybackRate returns the current video playback multiplier (e.g. 1.0x)
func (m *StateManager) VideoPlaybackRate() float64 {

	defer m.readLock()()

	if m.controllers == nil || m.controllers.videoPlayer == nil {
		return 0.0
	}

	return m.controllers.videoPlayer.PlaybackSpeed()
}

// initializeControllers creates the speed, video, and BLE controllers
func (m *StateManager) initializeControllers() (*controllers, error) {

	m.mu.RLock()
	cfg := m.activeConfig
	m.mu.RUnlock()

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating and initializing controllers...")

	if cfg == nil {
		return nil, errNoActiveConfig
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating new speed controller...")

	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating new video controller...")

	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create video controller: %w", err)
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating new BLE controller...")

	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create BLE controller: %w", err)
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "all controllers created and initialized")

	return &controllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, nil
}

// connectBLE handles BLE scanning, connection, and service discovery
func (m *StateManager) connectBLE(ctrl *controllers, shutdownMgr *services.ShutdownManager) (bluetooth.Device, error) {

	ctx := *shutdownMgr.Context()

	// Scan for BLE peripheral
	scanResult, err := ctrl.bleController.ScanForBLEPeripheral(ctx)
	if err != nil {
		return bluetooth.Device{}, fmt.Errorf("BLE scan failed: %w", err)
	}

	m.mu.Lock()
	m.state = StateConnecting
	m.mu.Unlock()

	// Connect to peripheral
	device, err := ctrl.bleController.ConnectToBLEPeripheral(ctx, scanResult)
	if err != nil {
		return bluetooth.Device{}, fmt.Errorf("BLE connection failed: %w", err)
	}

	m.mu.Lock()
	m.state = StateConnected
	m.mu.Unlock()

	// Get battery service
	batteryServices, err := ctrl.bleController.BatteryService(ctx, &device)
	if err != nil {
		return bluetooth.Device{}, fmt.Errorf("failed to get battery service: %w", err)
	}

	// Get battery level
	if err = ctrl.bleController.BatteryLevel(ctx, batteryServices); err != nil {
		return bluetooth.Device{}, fmt.Errorf("failed to get battery level: %w", err)
	}

	// Get CSC services
	cscServices, err := ctrl.bleController.CSCServices(ctx, &device)
	if err != nil {
		return bluetooth.Device{}, fmt.Errorf("failed to get CSC services: %w", err)
	}

	// Get CSC characteristics
	if err := ctrl.bleController.CSCCharacteristics(ctx, cscServices); err != nil {
		return bluetooth.Device{}, fmt.Errorf("failed to get CSC characteristics: %w", err)
	}

	return device, nil
}

// startServices launches BLE and video services in background goroutines
func (m *StateManager) startServices(ctrl *controllers, shutdownMgr *services.ShutdownManager) {

	m.runService(shutdownMgr, "BLE", func(ctx context.Context) error {
		return ctrl.bleController.BLEUpdates(ctx, ctrl.speedController)
	})

	m.runService(shutdownMgr, "video", func(ctx context.Context) error {
		return ctrl.videoPlayer.StartPlayback(ctx, ctrl.speedController)
	})

	logger.Debug(*shutdownMgr.Context(), logger.APP, "BLE and video services started")

}

// cleanupStartFailure handles cleaning manager state when session startup fails
func (m *StateManager) cleanupStartFailure(shutdownMgr *services.ShutdownManager) {

	logger.Debug(logger.BackgroundCtx, logger.APP, "resetting controllers and session state...")

	m.mu.Lock()
	m.PendingStart = false
	m.state = StateLoaded
	m.controllers = nil
	m.shutdownMgr = nil
	m.activeConfig = nil

	m.mu.Unlock()

	// ensure the shutdown manager... uh... shuts down
	if shutdownMgr != nil {
		shutdownMgr.Shutdown()
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "controllers and session state reset complete")

}

// runService helper to launch a service with standard error handling and logging
func (m *StateManager) runService(shutdownMgr *services.ShutdownManager, service string, action func(context.Context) error) {

	logger.Debug(*shutdownMgr.Context(), logger.APP, fmt.Sprintf("starting %s service goroutine", service))

	shutdownMgr.Run(func(ctx context.Context) error {
		logger.Debug(ctx, logger.APP, service+" service starting")

		err := action(ctx)

		// If this goroutine fails, we must reset the state and clean up resources
		if err != nil && !errors.Is(err, context.Canceled) {
			m.mu.Lock()
			// Only update if we were previously running (avoids clobbering other states if raced)
			if m.state == StateRunning {
				m.state = StateError
				m.errorMsg = fmt.Sprintf("%s service failed: %v", service, err)
			}
			// Rest resources state
			m.controllers = nil
			m.activeConfig = nil
			m.mu.Unlock()
		}

		return fmt.Errorf(errFormat, service+" service failed", err)
	})

}
