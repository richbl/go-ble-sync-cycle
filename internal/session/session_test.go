package session

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

var errTest = errors.New("test error message")

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
	configPath := newTestConfigFile(t)

	// Test loading a valid config
	loadSession(t, configPath, mgr)

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
	if mgr.LoadedConfigPath() != configPath {
		t.Errorf("LoadSession() path = %v, want %v", mgr.LoadedConfigPath(), configPath)
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
	err := mgr.LoadTargetSession(filepath.Join(t.TempDir(), "nonexistent.toml"))
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
	loadSession(t, newTestConfigFile(t), mgr)

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
	if mgr.LoadedConfigPath() != "" {
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
	loadSession(t, newTestConfigFile(t), mgr)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for range 10 {
		wg.Go(func() {
			for range iterations {
				_ = mgr.SessionState()
				_ = mgr.Config()
				_ = mgr.LoadedConfigPath()
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
	firstPath := newTestConfigFile(t)
	loadSession(t, firstPath, mgr)

	if mgr.LoadedConfigPath() != firstPath {
		t.Errorf("First load path = %v, want %v", mgr.LoadedConfigPath(), firstPath)
	}

	// Load second session (a different file, which simulates switching)
	secondPath := newTestConfigFile(t)
	loadSession(t, secondPath, mgr)

	if mgr.LoadedConfigPath() != secondPath {
		t.Errorf("Second load path = %v, want %v", mgr.LoadedConfigPath(), secondPath)
	}

	// Verify state is still Loaded
	if mgr.SessionState() != StateLoaded {
		t.Errorf("State after second load = %v, want %v", mgr.SessionState(), StateLoaded)
	}

}

// newTestConfigFile writes a valid session config (backed by a placeholder video file) to a
// temporary directory, and returns the config file path
//
// Config validation only checks that the video file exists, so an empty placeholder is enough
// here: session tests never need a playable video. Because the config uses absolute paths, the
// tests don't depend on the working directory, or on files in any other package
func newTestConfigFile(t *testing.T) string {

	t.Helper()

	dir := t.TempDir()

	videoPath := filepath.Join(dir, "test_video.mp4")
	if err := os.WriteFile(videoPath, nil, 0600); err != nil {
		t.Fatalf("failed to create placeholder video file: %v", err)
	}

	configPath := filepath.Join(dir, "test_session.toml")
	if err := config.Save(configPath, config.NewDefault(videoPath), config.GetVersion()); err != nil {
		t.Fatalf("failed to save test session config: %v", err)
	}

	return configPath
}

// loadSession is a helper function that loads a valid session configuration
func loadSession(t *testing.T, configPath string, mgr *StateManager) {

	t.Helper()

	if err := mgr.LoadTargetSession(configPath); err != nil {
		t.Fatalf("LoadTargetSession(%q) unexpected error: %v", configPath, err)
	}

}
