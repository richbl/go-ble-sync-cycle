// Package session manages BSC session lifecycle and state
package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/ble"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
	"github.com/richbl/go-ble-sync-cycle/internal/speed"
	"github.com/richbl/go-ble-sync-cycle/internal/video"
	"tinygo.org/x/bluetooth"
)

const (
	errFormat = "%v: %w"
)

var (
	errNoSessionLoaded = errors.New("no session loaded")
)

// State represents the current state of a session
type State int

const (
	StateIdle State = iota
	StateLoaded
	StateConnecting
	StateConnected
	StateRunning
	StatePaused
	StateError
)

// String returns a human-readable representation of the state
func (s State) String() string {

	return [...]string{
		"Idle",
		"Loaded",
		"Connecting",
		"Connected",
		"Running",
		"Paused",
		"Error",
	}[s]
}

// StateManager coordinates session lifecycle and state
type StateManager struct {
	config       *config.Config
	controllers  *controllers
	shutdownMgr  *services.ShutdownManager
	configPath   string
	errorMsg     string
	state        State
	mu           sync.RWMutex
	PendingStart bool
}

// controllers holds the application component controllers
type controllers struct {
	speedController *speed.Controller
	videoPlayer     *video.PlaybackController
	bleController   *ble.Controller
	bleDevice       bluetooth.Device
}

// NewManager creates a new session manager in Idle state
func NewManager() *StateManager {

	return &StateManager{
		state: StateIdle,
	}
}

// LoadSession loads and validates a session configuration file
func (m *StateManager) LoadSession(configPath string) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	// Load and validate the configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		m.state = StateError
		m.errorMsg = err.Error()

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update state
	m.config = cfg
	m.configPath = configPath
	m.state = StateLoaded
	m.errorMsg = ""

	// Set logging level
	if cfg.App.LogLevel != "" {
		logger.SetLogLevel(cfg.App.LogLevel)
	}

	return nil
}

// SessionState returns the current session state (thread-safe)
func (m *StateManager) SessionState() State {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state
}

// Config returns a copy of the current configuration (thread-safe)
func (m *StateManager) Config() *config.Config {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return nil
	}

	// Return a pointer to the config (caller should treat as read-only)
	return m.config
}

// ConfigPath returns the path to the loaded configuration file
func (m *StateManager) ConfigPath() string {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.configPath
}

// ErrorMessage returns the last error message if state is StateError
func (m *StateManager) ErrorMessage() string {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.errorMsg
}

// SetState updates the session state (used by service controllers)
func (m *StateManager) SetState(newState State) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = newState
}

// SetError sets the error state with a message
func (m *StateManager) SetError(err error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = StateError
	if err != nil {
		m.errorMsg = err.Error()
	} else {
		m.errorMsg = ""
	}

}

// Reset clears the session back to Idle state
func (m *StateManager) Reset() {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = StateIdle
	m.config = nil
	m.configPath = ""
	m.errorMsg = ""

}

// IsLoaded returns true if a session is currently loaded
func (m *StateManager) IsLoaded() bool {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config != nil && m.state != StateIdle
}

// IsRunning returns true if services are currently running
func (m *StateManager) IsRunning() bool {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state == StateRunning
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
	m.state = StateLoaded
	m.PendingStart = false
	m.mu.Unlock()

	if shutdownMgr == nil && !wasPending {
		return fmt.Errorf(errFormat, "no active session to stop", nil)
	}

	if wasPending {
		logger.Info(*shutdownMgr.Context(), logger.BLE, "stop requested, canceling pending BLE setup...")
	} else {
		logger.Info(*shutdownMgr.Context(), logger.BLE, "stop requested, canceling active session...")
	}

	// Trigger shutdown (cancels ctx, waits wg—like Ctrl+C)
	fmt.Fprint(os.Stdout, "\r") // Clear the ^C character from the terminal line

	if shutdownMgr != nil {
		shutdownMgr.Shutdown()
	}

	// Clear resources under lock
	m.mu.Lock()
	m.controllers = nil
	m.shutdownMgr = nil
	m.mu.Unlock()

	// Emulate CLI cleanup: stop any ongoing scan under mutex
	ble.AdapterMu.Lock()
	defer ble.AdapterMu.Unlock()

	if err := bluetooth.DefaultAdapter.StopScan(); err != nil {
		logger.Warn(*shutdownMgr.Context(), logger.BLE, fmt.Sprintf("failed to stop current BLE scan: %v", err))
	} else {
		logger.Info(*shutdownMgr.Context(), logger.BLE, "BLE scan stopped")
	}

	if wasPending {
		logger.Info(*shutdownMgr.Context(), logger.APP, "stopped pending session startup")
	} else {
		logger.Info(*shutdownMgr.Context(), logger.APP, "session stopped")
	}

	return nil
}

// BatteryLevel returns the current battery level from the BLE controller
func (m *StateManager) BatteryLevel() byte {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.controllers != nil && m.controllers.bleController != nil {
		return m.controllers.bleController.BatteryLevelLast()
	}

	return 0 // Unknown (0%)
}

// CurrentSpeed returns the current smoothed speed from the speed controller
func (m *StateManager) CurrentSpeed() (float64, string) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Guard against nil controllers (session stopped or not started)
	if m.controllers == nil || m.controllers.speedController == nil || m.config == nil {
		return 0.0, ""
	}

	return m.controllers.speedController.SmoothedSpeed(), m.config.Speed.SpeedUnits
}

// VideoTimeRemaining returns the formatted time remaining string (HH:MM:SS)
func (m *StateManager) VideoTimeRemaining() string {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.controllers == nil || m.controllers.videoPlayer == nil {
		return "--:--:--"
	}

	timeStr, err := m.controllers.videoPlayer.TimeRemaining()
	if err != nil {
		return "--:--:--"
	}

	return timeStr
}

// VideoPlaybackRate returns the current video playback multiplier (e.g. 1.0x)
func (m *StateManager) VideoPlaybackRate() float64 {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.controllers == nil || m.controllers.videoPlayer == nil {
		return 0.0
	}

	return m.controllers.videoPlayer.PlaybackSpeed()
}

// Context returns the session's context (NEW for CLI mode)
func (m *StateManager) Context() context.Context {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shutdownMgr != nil {
		return *m.shutdownMgr.Context()
	}

	return logger.BackgroundCtx
}

// Wait blocks until the session completes or is interrupted (NEW for CLI mode)
func (m *StateManager) Wait() {

	m.mu.RLock()
	shutdownMgr := m.shutdownMgr
	m.mu.RUnlock()

	if shutdownMgr != nil {
		shutdownMgr.Wait()
	}
}

// prepareStart validates state and sets PendingStart/state to Connecting
func (m *StateManager) prepareStart() error {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, "exiting: no config")

		return fmt.Errorf(errFormat, errNoSessionLoaded, nil)
	}

	if m.state == StateError {
		logger.Debug(logger.BackgroundCtx, logger.APP, "reset from Error state to Loaded state")
		m.state = StateLoaded
	}

	if m.state != StateLoaded {
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("exiting: invalid state for start: %s", m.state))

		return fmt.Errorf(errFormat, fmt.Sprintf("session already started or in invalid state: %s", m.state), nil)
	}

	if m.controllers != nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, "exiting: controllers already exist")

		return fmt.Errorf(errFormat, "session already started", nil)
	}

	m.PendingStart = true
	m.state = StateConnecting
	logger.Debug(logger.BackgroundCtx, logger.APP, "set PendingStart=true, state=Connecting")

	return nil
}

// storeShutdownMgr stores the shutdown manager under lock
func (m *StateManager) storeShutdownMgr(s *services.ShutdownManager) {

	m.mu.Lock()
	m.shutdownMgr = s
	logger.Debug(logger.BackgroundCtx, logger.APP, "shutdownMgr stored")
	m.mu.Unlock()

}

// cleanupStartFailure handles cleaning manager state when session startup fails
func (m *StateManager) cleanupStartFailure(shutdownMgr *services.ShutdownManager) {

	m.mu.Lock()
	m.PendingStart = false
	m.state = StateLoaded
	m.controllers = nil
	m.shutdownMgr = nil
	m.mu.Unlock()

	// ensure the shutdown manager is shut down (cancels any ongoing operations)
	if shutdownMgr != nil {
		shutdownMgr.Shutdown()
	}

}

// initializeControllers creates the speed, video, and BLE controllers
func (m *StateManager) initializeControllers() (*controllers, error) {

	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating speed controller...")
	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating video controller...")
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create video controller: %w", err)
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, "creating BLE controller...")
	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to create BLE controller: %w", err)
	}

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

	// Run BLE service
	shutdownMgr.Run(func(ctx context.Context) error {
		return ctrl.bleController.BLEUpdates(ctx, ctrl.speedController)
	})

	// Run video service
	shutdownMgr.Run(func(ctx context.Context) error {
		return ctrl.videoPlayer.Start(ctx, ctrl.speedController)
	})

}
