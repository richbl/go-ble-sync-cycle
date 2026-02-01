package session

import (
	"context"
	"errors"
	"fmt"
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
	errNoActiveConfig        = errors.New("cannot initialize controllers: no active configuration")
	errNoActiveSession       = errors.New("no active session to stop")
	errInitializeControllers = errors.New("failed to initialize controllers")
	errBLEConnectionFailed   = errors.New("failed to connect to BLE device")
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

	// Confirm start state, otherwise... why are we here?
	if err := m.prepareStart(); err != nil {
		return err
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "session startup sequence starting...")

	shutdownMgr := services.NewShutdownManager(30 * time.Second)
	shutdownMgr.Start()
	m.storeShutdownMgr(shutdownMgr)

	setupDone := make(chan error, 1)

	// Wrap connection phase in a managed WaitGroup to ensure clean shutdown
	shutdownMgr.Run(func(ctx context.Context) error {

		err := m.performSessionStartup(ctx, shutdownMgr)
		setupDone <- err

		return err
	})

	// Wait for connection success, internal failure, or user cancellation
	select {

	case err := <-setupDone:

		if err != nil {
			m.cleanupStartFailure(shutdownMgr)

			return err
		}

		logger.Debug(logger.BackgroundCtx, logger.APP, "session startup sequence completed")

		return nil

	case <-(*shutdownMgr.Context()).Done():
		m.cleanupStartFailure(shutdownMgr)

		return context.Canceled
	}

}

// performSessionStartup handles the initialization and connection logic for a session
func (m *StateManager) performSessionStartup(ctx context.Context, shutdownMgr *services.ShutdownManager) error {

	logger.Debug(ctx, logger.APP, "initializing controllers...")

	controllers, err := m.initializeControllers(ctx)
	if err != nil {
		logger.Error(ctx, logger.APP, fmt.Sprintf("controllers init failed: %v", err))

		return fmt.Errorf(errFormat, errInitializeControllers, err)
	}

	// Check if user clicked 'Stop' during init
	if err := ctx.Err(); err != nil {
		return fmt.Errorf(errFormat, errInitializeControllers, err)
	}

	logger.Debug(ctx, logger.APP, "controllers initialized OK")
	logger.Debug(ctx, logger.APP, "establishing connection to BLE peripheral...")

	// Connect to the BLE peripheral
	device, err := m.connectBLE(ctx, controllers)
	if err != nil {
		logger.Error(ctx, logger.APP, fmt.Sprintf("BLE connect failed: %v", err))

		return fmt.Errorf(errFormat, errBLEConnectionFailed, err)
	}

	controllers.bleDevice = device

	logger.Debug(ctx, logger.APP, "BLE peripheral now connected")

	m.mu.Lock()
	m.controllers = controllers
	m.state = StateRunning
	m.PendingStart = false
	m.mu.Unlock()

	logger.Debug(ctx, logger.APP, "starting services...")
	m.startServices(ctx, controllers, shutdownMgr)
	logger.Debug(ctx, logger.APP, "services started")

	return nil
}

// StopSession stops all services and cleans up controllers
func (m *StateManager) StopSession() error {

	m.mu.Lock()

	// Capture the manager instance we are about to stop
	targetMgr := m.shutdownMgr
	wasPending := m.PendingStart

	// Log the release of specific controller IDs before we destroy the manager object
	m.logControllersRelease(targetMgr)

	// Reset state
	m.state = StateLoaded
	m.PendingStart = false

	// Null the StateManager fields only if they still point to the manager we are stopping
	if m.shutdownMgr == targetMgr {
		m.controllers = nil
		m.shutdownMgr = nil
		m.activeConfig = nil
	}

	m.mu.Unlock()

	// If there's nothing to stop, return
	if targetMgr == nil && !wasPending {
		return errNoActiveSession
	}

	// Determine the log context (use the target manager's context if available)
	ctx := logger.BackgroundCtx

	if targetMgr != nil {
		ctx = *targetMgr.Context()
	}

	if wasPending {
		logger.Debug(ctx, logger.APP, "stop requested, canceling pending session setup...")
	} else {
		logger.Debug(ctx, logger.APP, "stop requested, canceling active session...")
	}

	// Finally, we can now stop the session object
	if targetMgr != nil {
		targetMgr.Shutdown()
	}

	if wasPending {
		logger.Debug(ctx, logger.APP, "stopped pending session startup")
	} else {
		logger.Debug(ctx, logger.APP, "active session stopped")
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
		logger.Debug(ctx, logger.BLE, fmt.Sprintf("releasing BLE controller object (id:%04d)", m.controllers.bleController.InstanceID))
	}
	if m.controllers.speedController != nil {
		logger.Debug(ctx, logger.SPEED, fmt.Sprintf("releasing speed controller object (id:%04d)", m.controllers.speedController.InstanceID))
	}
	if m.controllers.videoPlayer != nil {
		logger.Debug(ctx, logger.VIDEO, fmt.Sprintf("releasing video controller object (id:%04d)", m.controllers.videoPlayer.InstanceID))
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
func (m *StateManager) initializeControllers(ctx context.Context) (*controllers, error) {

	m.mu.RLock()
	cfg := m.activeConfig
	m.mu.RUnlock()

	logger.Debug(ctx, logger.APP, "creating and initializing controllers...")

	if cfg == nil {
		return nil, errNoActiveConfig
	}

	logger.Debug(ctx, logger.APP, "creating new speed controller...")
	speedController := speed.NewSpeedController(ctx, cfg.Speed.SmoothingWindow)
	logger.Debug(ctx, logger.APP, "creating new video controller...")

	videoPlayer, err := video.NewPlaybackController(ctx, cfg.Video, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create video controller: %w", err)
	}

	logger.Debug(ctx, logger.APP, "creating new BLE controller...")
	bleController, err := ble.NewBLEController(ctx, cfg.BLE, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create BLE controller: %w", err)
	}

	logger.Debug(ctx, logger.APP, "all controllers created and initialized")

	return &controllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, nil
}

// connectBLE handles BLE scanning, connection, and service discovery
func (m *StateManager) connectBLE(ctx context.Context, ctrl *controllers) (bluetooth.Device, error) {

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
func (m *StateManager) startServices(ctx context.Context, ctrl *controllers, shutdownMgr *services.ShutdownManager) {

	m.runService(ctx, shutdownMgr, "BLE", func(ctx context.Context) error {
		return ctrl.bleController.BLEUpdates(ctx, ctrl.speedController)
	})

	m.runService(ctx, shutdownMgr, "video", func(ctx context.Context) error {
		return ctrl.videoPlayer.StartPlayback(ctx, ctrl.speedController)
	})

	logger.Debug(ctx, logger.APP, "BLE and video services started")

}

// cleanupStartFailure handles cleaning manager state when session startup fails
func (m *StateManager) cleanupStartFailure(shutdownMgr *services.ShutdownManager) {

	m.mu.Lock()

	// Confirm that the shutdown manager being cleaned up is the one we currently hold
	isCurrent := (m.shutdownMgr == shutdownMgr)

	if isCurrent {
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("resetting state for current shutdown manager (id:%04d)", shutdownMgr.InstanceID))
		m.PendingStart = false
		m.state = StateLoaded
		m.controllers = nil
		m.shutdownMgr = nil
		m.activeConfig = nil
	}
	m.mu.Unlock()

	// If no shutdown manager to clean up, go home
	if shutdownMgr == nil {
		return
	}

	if (*shutdownMgr.Context()).Err() == nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("shutting down shutdown manager (id:%04d)", shutdownMgr.InstanceID))
		shutdownMgr.Shutdown()
	}

	if isCurrent {
		logger.Debug(logger.BackgroundCtx, logger.APP, "controllers and session state reset complete")
	}

}

// runService helper to launch a service with standard error handling and logging
func (m *StateManager) runService(ctx context.Context, shutdownMgr *services.ShutdownManager, service string, action func(context.Context) error) {

	logger.Debug(ctx, logger.APP, fmt.Sprintf("starting %s service goroutine", service))

	shutdownMgr.Run(func(ctx context.Context) error {

		logger.Debug(ctx, logger.APP, service+" service starting")

		err := action(ctx)

		// If this goroutine fails, we reset the state and clean up resources
		if err != nil && !errors.Is(err, context.Canceled) {

			m.mu.Lock()

			// Only update if we were previously running
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
