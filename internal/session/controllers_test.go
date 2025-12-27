package session

import (
	"testing"
)

// TestBatteryLevel tests the BatteryLevel accessor
func TestBatteryLevel(t *testing.T) {

	mgr := NewManager()

	// Test with no controllers (should return 0)
	level := mgr.BatteryLevel()
	if level != 0 {
		t.Errorf("BatteryLevel() = %v, want 0 when no controllers exist", level)
	}

	// TODO: Add tests with mock controllers once BLE controller is mockable

}

// TestCurrentSpeed tests the CurrentSpeed accessor
func TestCurrentSpeed(t *testing.T) {

	mgr := NewManager()

	// Test with no controllers (should return 0.0, "")
	speed, units := mgr.CurrentSpeed()
	if speed != 0.0 {
		t.Errorf("CurrentSpeed() speed = %v, want 0.0 when no controllers exist", speed)
	}
	if units != "" {
		t.Errorf("CurrentSpeed() units = %v, want empty string when no controllers exist", units)
	}

	// TODO: Add tests with mock controllers once speed controller is mockable

}

// TestVideoTimeRemaining tests the VideoTimeRemaining accessor
func TestVideoTimeRemaining(t *testing.T) {

	mgr := NewManager()

	// Test with no controllers (should return "--:--:--")
	timeStr := mgr.VideoTimeRemaining()
	if timeStr != "--:--:--" {
		t.Errorf("VideoTimeRemaining() = %v, want '--:--:--' when no controllers exist", timeStr)
	}

	// TODO: Add tests with mock video controller once it's mockable

}

// TestVideoPlaybackRate tests the VideoPlaybackRate accessor
func TestVideoPlaybackRate(t *testing.T) {

	mgr := NewManager()

	// Test with no controllers (should return 0.0)
	rate := mgr.VideoPlaybackRate()
	if rate != 0.0 {
		t.Errorf("VideoPlaybackRate() = %v, want 0.0 when no controllers exist", rate)
	}

	// TODO: Add tests with mock video controller once it's mockable

}

// TestStartSession tests the StartSession function
func TestStartSession(t *testing.T) {

	t.Skip("StartSession requires BLE hardware and is tested via integration tests")

	// TODO: Add unit tests with mock controllers
	// This would require:
	// 1. Mock BLE controller
	// 2. Mock video controller
	// 3. Mock speed controller
	// 4. Mock shutdown manager

}

// TestStopSession tests the StopSession function
func TestStopSession(t *testing.T) {

	mgr := NewManager()

	// Test stopping when no session is active
	err := mgr.StopSession()
	if err == nil {
		t.Error("StopSession() should return error when no session is active")
	}

	// TODO: Add tests with active session once controllers are mockable

}
