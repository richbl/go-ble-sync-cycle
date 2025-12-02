package session

import (
	"fmt"
	"sync"
	"testing"
)

// TestNewManager tests the creation of a new session manager
func TestNewManager(t *testing.T) {

	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if mgr.GetState() != StateIdle {
		t.Errorf("NewManager() state = %v, want %v", mgr.GetState(), StateIdle)
	}

	if mgr.GetConfig() != nil {
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
	err := mgr.LoadSession("../config/config_test.toml")
	if err != nil {
		t.Fatalf("LoadSession() unexpected error: %v", err)
	}

	// Verify state changed to Loaded
	if mgr.GetState() != StateLoaded {
		t.Errorf("LoadSession() state = %v, want %v", mgr.GetState(), StateLoaded)
	}

	// Verify config is loaded
	cfg := mgr.GetConfig()
	if cfg == nil {
		t.Fatal("LoadSession() config should not be nil")
	}

	// Verify path is stored
	expectedPath := "../config/config_test.toml"

	if mgr.GetConfigPath() != expectedPath {
		t.Errorf("LoadSession() path = %v, want %v", mgr.GetConfigPath(), expectedPath)
	}

	// Verify IsLoaded returns true
	if !mgr.IsLoaded() {
		t.Error("LoadSession() IsLoaded() should be true after loading")
	}

	// Verify no error message
	if mgr.GetError() != "" {
		t.Errorf("LoadSession() error message should be empty, got: %v", mgr.GetError())
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
	if mgr.GetState() != StateError {
		t.Errorf("LoadSession() state = %v, want %v", mgr.GetState(), StateError)
	}

	// Verify error message is set
	if mgr.GetError() == "" {
		t.Error("LoadSession() error message should not be empty")
	}

	// Verify config is still nil
	if mgr.GetConfig() != nil {
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

		if mgr.GetState() != expected {
			t.Errorf("SetState() state = %v, want %v", mgr.GetState(), expected)
		}

	}

}

// TestSetError tests error state management
func TestSetError(t *testing.T) {

	mgr := NewManager()

	// Test with actual error
	testErr := fmt.Errorf("test error message")
	mgr.SetError(testErr)

	if mgr.GetState() != StateError {
		t.Errorf("SetError() state = %v, want %v", mgr.GetState(), StateError)
	}

	if mgr.GetError() != testErr.Error() {
		t.Errorf("SetError() error = %v, want %v", mgr.GetError(), testErr.Error())
	}

	// Test with nil error (should not panic)
	mgr.SetError(nil)

	if mgr.GetState() != StateError {
		t.Errorf("SetError(nil) state = %v, want %v", mgr.GetState(), StateError)
	}

}

// TestReset tests resetting the manager back to idle state
func TestReset(t *testing.T) {

	mgr := NewManager()

	// Load a session first
	err := mgr.LoadSession("../config/config_test.toml")
	if err != nil {
		t.Fatalf("LoadSession() unexpected error: %v", err)
	}

	// Verify session is loaded
	if !mgr.IsLoaded() {
		t.Fatal("Session should be loaded before reset")
	}

	// Reset the manager
	mgr.Reset()

	// Verify state is Idle
	if mgr.GetState() != StateIdle {
		t.Errorf("Reset() state = %v, want %v", mgr.GetState(), StateIdle)
	}

	// Verify config is cleared
	if mgr.GetConfig() != nil {
		t.Error("Reset() config should be nil")
	}

	// Verify path is cleared
	if mgr.GetConfigPath() != "" {
		t.Error("Reset() path should be empty")
	}

	// Verify error message is cleared
	if mgr.GetError() != "" {
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
	err := mgr.LoadSession("../config/config_test.toml")
	if err != nil {
		t.Fatalf("LoadSession() unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for range 10 {
		wg.Go(func() {
			for j := 0; j < iterations; j++ {
				_ = mgr.GetState()
				_ = mgr.GetConfig()
				_ = mgr.GetConfigPath()
				_ = mgr.GetError()
				_ = mgr.IsLoaded()
			}
		})
	}

	// Concurrent state changes
	for range 5 {
		wg.Go(func() {
			for j := 0; j < iterations; j++ {
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
	err := mgr.LoadSession("../config/config_test.toml")
	if err != nil {
		t.Fatalf("LoadSession() first load failed: %v", err)
	}

	firstPath := mgr.GetConfigPath()

	// Load second session (same file, but simulates switching)
	err = mgr.LoadSession("../config/config_test.toml")
	if err != nil {
		t.Fatalf("LoadSession() second load failed: %v", err)
	}

	secondPath := mgr.GetConfigPath()

	// Verify both loads succeeded
	if firstPath != secondPath {
		t.Errorf("Paths differ: first=%v, second=%v", firstPath, secondPath)
	}

	// Verify state is still Loaded
	if mgr.GetState() != StateLoaded {
		t.Errorf("State after second load = %v, want %v", mgr.GetState(), StateLoaded)
	}

}
