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
	errFormat    = "%v: %w"
	errFormatRev = "%w: %v"
)

// Error definitions
var (
	errNoSessionLoaded       = errors.New("no session loaded")
	errSessionAlreadyStarted = errors.New("session already started")
	errInvalidState          = errors.New("invalid state for start")
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
	activeConfig *config.Config // The "currently running" config

	loadedConfig     *config.Config // The "loaded" (but not running) config
	loadedConfigPath string

	editConfig     *config.Config // The "getting edited" config
	editConfigPath string

	controllers  *controllers
	shutdownMgr  *services.ShutdownManager
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

// LoadTargetSession loads (or reloads) a session configuration for execution
func (m *StateManager) LoadTargetSession(configPath string) error {

	defer m.writeLock()()

	cfg, err := config.Load(configPath)
	if err != nil {

		if m.state != StateRunning && m.state != StatePaused && m.state != StateConnected {
			m.state = StateError
		}

		m.errorMsg = err.Error()

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	m.loadedConfig = cfg
	m.loadedConfigPath = configPath

	// When loading a session to run, we also queue it for editing
	m.editConfig = cfg
	m.editConfigPath = configPath

	m.errorMsg = ""
	if m.state == StateIdle || m.state == StateError {
		m.state = StateLoaded
	}

	if cfg.App.LogLevel != "" {
		logger.SetLogLevel(cfg.App.LogLevel)
	}

	return nil
}

// LoadEditSession loads a session configuration specifically for editing
func (m *StateManager) LoadEditSession(configPath string) error {

	defer m.writeLock()()

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration for edit: %w", err)
	}

	m.editConfig = cfg
	m.editConfigPath = configPath

	return nil
}

// UpdateLoadedSession updates the loaded session configuration
func (m *StateManager) UpdateLoadedSession(cfg *config.Config, path string) error {

	defer m.writeLock()()

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Update the loaded config
	m.editConfig = cfg
	m.editConfigPath = path

	// If the path matches the loaded config path, update the loaded config too
	if m.loadedConfigPath == path {
		m.loadedConfig = cfg

		// Set the state to Loaded if we were in Error or Idle state
		if m.state == StateError || m.state == StateIdle {
			m.state = StateLoaded
			m.errorMsg = ""
		}

	}

	return nil
}

// SessionState returns the current session state
func (m *StateManager) SessionState() State {

	defer m.readLock()()

	return m.state
}

// Config returns a copy of the current editing configuration
func (m *StateManager) Config() *config.Config {

	defer m.readLock()()

	return m.editConfig
}

// ActiveConfig returns the configuration of the currently running/loaded session
func (m *StateManager) ActiveConfig() *config.Config {

	defer m.readLock()()

	// If running, return an immutable snapshot
	if m.activeConfig != nil {
		return m.activeConfig
	}

	// If a session is loaded to run, return it
	if m.loadedConfig != nil {
		return m.loadedConfig
	}

	// Fallback to editConfig (default behavior)
	return m.editConfig
}

// EditConfigPath returns the path to the configuration currently being edited
func (m *StateManager) EditConfigPath() string {

	defer m.readLock()()

	return m.editConfigPath
}

// LoadedConfigPath returns the path to the loaded/running configuration
func (m *StateManager) LoadedConfigPath() string {

	defer m.readLock()()

	return m.loadedConfigPath
}

// ErrorMessage returns the last error message if state is StateError
func (m *StateManager) ErrorMessage() string {

	defer m.readLock()()

	return m.errorMsg
}

// SetState updates the session state (used by service controllers)
func (m *StateManager) SetState(newState State) {

	defer m.writeLock()()
	m.state = newState

}

// SetError sets the error state with a message
func (m *StateManager) SetError(err error) {

	defer m.writeLock()()

	m.state = StateError
	if err != nil {
		m.errorMsg = err.Error()
	} else {
		m.errorMsg = ""
	}

}

// Reset clears the session back to Idle state
func (m *StateManager) Reset() {

	defer m.writeLock()()

	m.state = StateIdle
	m.editConfig = nil
	m.loadedConfig = nil
	m.activeConfig = nil
	m.editConfigPath = ""
	m.loadedConfigPath = ""
	m.errorMsg = ""

}

// IsLoaded returns true if a session is currently loaded
func (m *StateManager) IsLoaded() bool {

	defer m.readLock()()

	return (m.loadedConfig != nil || m.editConfig != nil) && m.state != StateIdle
}

// IsRunning returns true if services are currently running
func (m *StateManager) IsRunning() bool {

	defer m.readLock()()

	return m.state == StateRunning
}

// Context returns the session's context
func (m *StateManager) Context() context.Context {

	defer m.readLock()()

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

	defer m.writeLock()()

	if m.editConfig == nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, "exiting: no config")

		return errNoSessionLoaded
	}

	// Create a snapshot of the config
	switch {
	case m.loadedConfig != nil:
		m.activeConfig = m.loadedConfig
	case m.editConfig != nil:
		m.activeConfig = m.editConfig
	default:

		return errNoSessionLoaded
	}

	if m.state == StateError {
		logger.Debug(logger.BackgroundCtx, logger.APP, "reset from Error state to Loaded state")
		m.state = StateLoaded
	}

	if m.state != StateLoaded {
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("exiting: invalid state for start: %s", m.state))

		return fmt.Errorf(errFormatRev, errInvalidState, m.state)
	}

	if m.controllers != nil {
		logger.Debug(logger.BackgroundCtx, logger.APP, "exiting: controllers already exist")

		return errSessionAlreadyStarted
	}

	m.PendingStart = true
	m.state = StateConnecting

	return nil
}

// storeShutdownMgr stores the shutdown manager under lock
func (m *StateManager) storeShutdownMgr(s *services.ShutdownManager) {

	m.mu.Lock()
	m.shutdownMgr = s
	logger.Debug(logger.BackgroundCtx, logger.APP, "ShutdownManager object state stored")
	m.mu.Unlock()

}

// readLock acquires a read lock and returns a function to release it
func (m *StateManager) readLock() func() {

	m.mu.RLock()

	return m.mu.RUnlock
}

// writeLock acquires a write lock and returns a function to release it
func (m *StateManager) writeLock() func() {

	m.mu.Lock()

	return m.mu.Unlock
}
