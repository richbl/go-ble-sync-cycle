package config

import (
	"os"
	"testing"
)

func TestLoadFile(t *testing.T) {
	// Create a test config file
	configFile := []byte(`
	[ble]
	sensor_uuid = "12345678-1234-1234-1234-123456789012"
	scan_timeout_secs = 10

	[speed]
	smoothing_window = 5
	speed_threshold = 10.0
	wheel_circumference_mm = 2000
	speed_units = "km/h"

	[video]
	file_path = "[VIDEO]path/to/video.mp4"
	update_interval_sec = 1
	speed_multiplier = 2.0
`)

	// Write the config file to a temporary file
	tmpFile, err := os.CreateTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(configFile)
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Load the config file
	cfg, err := LoadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Verify the config values
	if cfg.BLE.SensorUUID != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("BLE sensor UUID mismatch: expected %q, got %q", "12345678-1234-1234-1234-123456789012", cfg.BLE.SensorUUID)
	}
	if cfg.Speed.SmoothingWindow != 5 {
		t.Errorf("Speed smoothing window mismatch: expected %d, got %d", 5, cfg.Speed.SmoothingWindow)
	}
	if cfg.Video.FilePath != "[VIDEO]path/to/video.mp4" {
		t.Errorf("Video file path mismatch: expected %q, got %q", "[VIDEO]path/to/video.mp4", cfg.Video.FilePath)
	}
}

func TestLoadFile_InvalidConfig(t *testing.T) {
	// Create an invalid config file
	configFile := []byte(`
	[ble]
	sensor_uuid = invalid-uuid
`)

	// Write the config file to a temporary file
	tmpFile, err := os.CreateTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(configFile)
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Load the config file
	_, err = LoadFile(tmpFile.Name())
	if err == nil {
		t.Errorf("Expected error loading invalid config file, but got nil")
	}
}

func TestLoadFile_MissingConfigFile(t *testing.T) {
	// Try to load a non-existent config file
	_, err := LoadFile("non-existent-config-file.toml")
	if err == nil {
		t.Errorf("Expected error loading missing config file, but got nil")
	}
}
