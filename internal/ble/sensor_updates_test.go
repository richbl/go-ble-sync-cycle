package ble

import (
	"testing"

	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// init is called to set the log level
func init() {
	logger.Initialize("debug")
}

// TestCalculateSpeed tests the calculateSpeed function
func TestCalculateSpeed(t *testing.T) {

	wheelCircumferenceMM := 2100 // Example wheel circumference in mm
	speedUnitMultiplier := 1.0   // km/h

	sd := initSpeedData(wheelCircumferenceMM, speedUnitMultiplier)
	sd.wheelRevs = 2
	sd.wheelTime = 1024

	// First call: Initialization phase
	speed := sd.calculateSpeed()
	if speed != 0.0 {
		t.Errorf("Expected speed of 0.0 during initialization, got %v", speed)
	}

	// Update the wheel data for the second call
	sd.wheelRevs = 4
	sd.wheelTime = 2048

	// Second call: Speed calculation
	speed = sd.calculateSpeed()
	expectedSpeed := 15.12 // Expected speed in km/h

	if speed != expectedSpeed {
		t.Errorf("Expected speed %v, got %v", expectedSpeed, speed)
	}

}

// TestParseSpeedData tests the parseSpeedData function
func TestParseSpeedData(t *testing.T) {

	sd := &speedData{}
	data := []byte{0x01, 0x64, 0x00, 0x00, 0x00, 0x00, 0x04} // Example BLE data

	err := sd.parseSpeedData(data)
	if err != nil {
		t.Errorf("Failed to parse speed data: %v", err)
	}

	if sd.wheelRevs != 100 || sd.wheelTime != 1024 {
		t.Errorf("Parsed data mismatch: expected wheelRevs=100, wheelTime=1024; got wheelRevs=%v, wheelTime=%v", sd.wheelRevs, sd.wheelTime)
	}

}
