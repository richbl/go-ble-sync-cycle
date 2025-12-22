package session

import (
	"errors"
	"sync"
	"testing"

	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

var (
	configPath     = "../config/config_test.toml"
	errLoadSession = errors.New("LoadSession() unexpected error: %v")
	errTest        = errors.New("test error message")
)

// init is called to set the log level for tests
func init() {
	logger.Initialize("debug")
}

// TestNewManager tests the creation of a new session manager
func TestNewManager(t *testing.T) {

	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if mgr.SessionState() != StateIdle {
		t.Errorf("NewManager() state = %v, want %v", mgr.SessionState(), StateIdle)
	}

	if mgr.Config() != nil {
		t.Error("NewManager() config should be nil")
	}

	if mgr.IsLoaded() {
		t.Error("NewManager() IsLoaded() should be false")
	}

}

// TestLoadSession tests loading a valid session configuration
func TestLoadSession(t *testing.T) {

	mgr := NewManager()

	// Test loading a valid config
	err := mgr.LoadSession(configPath)
	if err != nil {
		t.Fatalf(errLoadSession.Error(), err)
	}

	// Verify state changed to Loaded
	if mgr.SessionState() != StateLoaded {
		t.Errorf("LoadSession() state = %v, want %v", mgr.SessionState(), StateLoaded)
	}

	// Verify config is loaded
	cfg := mgr.Config()
	if cfg == nil {
		t.Fatal("LoadSession() config should not be nil")
	}

	// Verify path is stored
	expectedPath := configPath

	if mgr.ConfigPath() != expectedPath {
		t.Errorf("LoadSession() path = %v, want %v", mgr.ConfigPath(), expectedPath)
	}

	// Verify IsLoaded returns true
	if !mgr.IsLoaded() {
		t.Error("LoadSession() IsLoaded() should be true after loading")
	}

	// Verify no error message
	if mgr.ErrorMessage() != "" {
		t.Errorf("LoadSession() error message should be empty, got: %v", mgr.ErrorMessage())
	}

}

// TestLoadSessionInvalidFile tests loading an invalid configuration
func TestLoadSessionInvalidFile(t *testing.T) {

	mgr := NewManager()

	// Test loading a non-existent file
	err := mgr.LoadSession("nonexistent.toml")
	if err == nil {
		t.Error("LoadSession() expected error for non-existent file")
	}

	// Verify state changed to Error
	if mgr.SessionState() != StateError {
		t.Errorf("LoadSession() state = %v, want %v", mgr.SessionState(), StateError)
	}

	// Verify error message is set
	if mgr.ErrorMessage() == "" {
		t.Error("LoadSession() error message should not be empty")
	}

	// Verify config is still nil
	if mgr.Config() != nil {
		t.Error("LoadSession() config should be nil after failed load")
	}

}

// TestStateTransitions tests valid state transitions
func TestStateTransitions(t *testing.T) {

	mgr := NewManager()

	// Test state progression
	states := []State{StateLoaded, StateConnecting, StateConnected, StateRunning}

	for _, expected := range states {
		mgr.SetState(expected)

		if mgr.SessionState() != expected {
			t.Errorf("SetState() state = %v, want %v", mgr.SessionState(), expected)
		}

	}

}

// TestSetError tests error state management
func TestSetError(t *testing.T) {

	mgr := NewManager()

	// Test with actual error
	mgr.SetError(errTest)

	if mgr.SessionState() != StateError {
		t.Errorf("SetError() state = %v, want %v", mgr.SessionState(), StateError)
	}

	if mgr.ErrorMessage() != errTest.Error() {
		t.Errorf("SetError() error = %v, want %v", mgr.ErrorMessage(), errTest.Error())
	}

	// Test with nil error (should not panic)
	mgr.SetError(nil)

	if mgr.SessionState() != StateError {
		t.Errorf("SetError(nil) state = %v, want %v", mgr.SessionState(), StateError)
	}

}

// TestReset tests resetting the manager back to idle state
func TestReset(t *testing.T) {

	mgr := NewManager()

	// Load a session first
	err := mgr.LoadSession(configPath)
	if err != nil {
		t.Fatalf(errLoadSession.Error(), err)
	}

	// Verify session is loaded
	if !mgr.IsLoaded() {
		t.Fatal("Session should be loaded before reset")
	}

	// Reset the manager
	mgr.Reset()

	// Verify state is Idle
	if mgr.SessionState() != StateIdle {
		t.Errorf("Reset() state = %v, want %v", mgr.SessionState(), StateIdle)
	}

	// Verify config is cleared
	if mgr.Config() != nil {
		t.Error("Reset() config should be nil")
	}

	// Verify path is cleared
	if mgr.ConfigPath() != "" {
		t.Error("Reset() path should be empty")
	}

	// Verify error message is cleared
	if mgr.ErrorMessage() != "" {
		t.Error("Reset() error message should be empty")
	}

	// Verify IsLoaded returns false
	if mgr.IsLoaded() {
		t.Error("Reset() IsLoaded() should be false")
	}

}

// TestConcurrentAccess tests thread-safety of the manager
func TestConcurrentAccess(t *testing.T) {

	mgr := NewManager()

	// Load a session first
	err := mgr.LoadSession(configPath)
	if err != nil {
		t.Fatalf(errLoadSession.Error(), err)
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for range 10 {
		wg.Go(func() {
			for range iterations {
				_ = mgr.SessionState()
				_ = mgr.Config()
				_ = mgr.ConfigPath()
				_ = mgr.ErrorMessage()
				_ = mgr.IsLoaded()
			}
		})
	}

	// Concurrent state changes
	for range 5 {
		wg.Go(func() {
			for range iterations {
				mgr.SetState(StateConnecting)
				mgr.SetState(StateConnected)
			}
		})
	}

	wg.Wait()
}

// TestStateString tests the String() method for State
func TestStateString(t *testing.T) {

	tests := []struct {
		state    State
		expected string
	}{
		{StateIdle, "Idle"},
		{StateLoaded, "Loaded"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateRunning, "Running"},
		{StatePaused, "Paused"},
		{StateError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {

			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %v, want %v", got, tt.expected)
			}

		})
	}

}

// TestLoadSessionMultipleTimes tests loading different sessions sequentially
func TestLoadSessionMultipleTimes(t *testing.T) {

	mgr := NewManager()

	// Load first session
	err := mgr.LoadSession(configPath)
	if err != nil {
		t.Fatalf("LoadSession() first load failed: %v", err)
	}

	firstPath := mgr.ConfigPath()

	// Load second session (same file, but simulates switching)
	err = mgr.LoadSession(configPath)
	if err != nil {
		t.Fatalf("LoadSession() second load failed: %v", err)
	}

	secondPath := mgr.ConfigPath()

	// Verify both loads succeeded
	if firstPath != secondPath {
		t.Errorf("Paths differ: first=%v, second=%v", firstPath, secondPath)
	}

	// Verify state is still Loaded
	if mgr.SessionState() != StateLoaded {
		t.Errorf("State after second load = %v, want %v", mgr.SessionState(), StateLoaded)
	}

}
