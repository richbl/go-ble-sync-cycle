package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
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
	editConfig   *config.Config // The "editing" config (mutable)
	activeConfig *config.Config // The "currently running" config (immutable)
	controllers  *controllers
	shutdownMgr  *services.ShutdownManager
	configPath   string
	errorMsg     string
	state        State
	mu           sync.RWMutex
	PendingStart bool
}

// NewManager creates a new session manager in Idle state
func NewManager() *StateManager {
	return &StateManager{
		state: StateIdle,
	}
}

// LoadSession loads and validates a session configuration file into editConfig
func (m *StateManager) LoadSession(configPath string) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	// Load and validate the configuration from disk
	cfg, err := config.Load(configPath)
	if err != nil {
		// Only set error state if we aren't currently running a valid session
		if m.state != StateRunning && m.state != StatePaused && m.state != StateConnected {
			m.state = StateError
		}
		m.errorMsg = err.Error()

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update the editing config
	m.editConfig = cfg
	m.configPath = configPath
	m.errorMsg = ""

	// Do not downgrade state to 'Loaded' if we are currently 'Running'
	if m.state == StateIdle || m.state == StateError {
		m.state = StateLoaded
	}

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

// Config returns a copy of the current editing configuration (thread-safe)
func (m *StateManager) Config() *config.Config {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.editConfig
}

// ActiveConfig returns the configuration of the currently running session
func (m *StateManager) ActiveConfig() *config.Config {

	m.mu.RLock()
	defer m.mu.RUnlock()

	// If running, return the snapshot
	if m.activeConfig != nil {
		return m.activeConfig
	}

	// Fallback to editConfig (e.g., UI display before start)
	return m.editConfig
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
	m.editConfig = nil
	m.activeConfig = nil
	m.configPath = ""
	m.errorMsg = ""

}

// IsLoaded returns true if a session is currently loaded
func (m *StateManager) IsLoaded() bool {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.editConfig != nil && m.state != StateIdle
}

// IsRunning returns true if services are currently running
func (m *StateManager) IsRunning() bool {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state == StateRunning
}

// Context returns the session's context
func (m *StateManager) Context() context.Context {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shutdownMgr != nil {
		return *m.shutdownMgr.Context()
	}

	return logger.BackgroundCtx
}

// Wait blocks until the session completes or is interrupted
func (m *StateManager) Wait() {

	m.mu.RLock()
	shutdownMgr := m.shutdownMgr
	m.mu.RUnlock()

	if shutdownMgr != nil {
		shutdownMgr.Wait()
	}

}

// prepareStart validates state and snapshots editConfig to activeConfig
func (m *StateManager) prepareStart() error {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.editConfig == nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, "exiting: no config")

		return fmt.Errorf(errFormat, errNoSessionLoaded, nil)
	}

	// Create a snapshot of the config (activeConfig is immutable)
	m.activeConfig = m.editConfig

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
	m.activeConfig = nil
	m.mu.Unlock()

	// ensure the shutdown manager... uh... shuts down
	if shutdownMgr != nil {
		shutdownMgr.Shutdown()
	}

}
