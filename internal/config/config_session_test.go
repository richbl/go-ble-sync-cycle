package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Error messages
var (
	errFailedToCreateTempFile = errors.New("Failed to create temp config file: %v")
)

// TestLoadSessionMetadataSuccess tests the successful loading of session metadata
func TestLoadSessionMetadataSuccess(t *testing.T) {
	t.Run("valid config with session title", func(t *testing.T) {
		configFile := "config_test.toml"
		expectedTitle := "Session Title"

		metadata, err := LoadSessionMetadata(configFile)

		if err != nil {
			t.Fatalf("LoadSessionMetadata() returned unexpected error: %v", err)
		}

		if metadata == nil {
			t.Fatal("LoadSessionMetadata() returned nil metadata")
		}

		if !metadata.IsValid {
			t.Error("LoadSessionMetadata() metadata.IsValid should be true for valid configs")
		}

		if metadata.Title != expectedTitle {
			t.Errorf("LoadSessionMetadata() Title = %v, want %v", metadata.Title, expectedTitle)
		}

		if metadata.FilePath != configFile {
			t.Errorf("LoadSessionMetadata() FilePath = %v, want %v", metadata.FilePath, configFile)
		}
	})
}

// TestLoadSessionMetadataErrors tests error handling in LoadSessionMetadata
func TestLoadSessionMetadataErrors(t *testing.T) {

	// Define test cases
	tests := []struct {
		name       string
		configFile string
	}{
		{
			name:       "non-existent file",
			configFile: "non_existent.toml",
		},
		{
			name:       "invalid config file",
			configFile: "invalid_config.toml",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := LoadSessionMetadata(tt.configFile)

			if err == nil {
				t.Errorf("LoadSessionMetadata() expected an error, but got nil")
				return
			}

			if metadata == nil {
				t.Fatal("LoadSessionMetadata() should return metadata even on error")
			}

			if metadata.IsValid {
				t.Error("LoadSessionMetadata() metadata.IsValid should be false for invalid configs")
			}

			if metadata.ErrorMsg == "" {
				t.Error("LoadSessionMetadata() metadata.ErrorMsg should not be empty for invalid configs")
			}
		})
	}

}

// TestLoadSessionMetadataWithEmptyTitle tests behavior when session_title is empty
func TestLoadSessionMetadataWithEmptyTitle(t *testing.T) {

	// Create a temporary config file with empty session_title
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_session.toml")

	configContent := `
[app]
  session_title = ""
  logging_level = "info"

[ble]
  sensor_bd_addr = "FA:46:1D:77:C8:E1"
  scan_timeout_secs = 30

[speed]
  speed_threshold = 0.25
  speed_units = "mph"
  smoothing_window = 5
  wheel_circumference_mm = 2155

[video]
  media_player = "mpv"
  file_path = "test_video.mp4"
  window_scale_factor = 1.0
  seek_to_position = "00:00"
  update_interval_secs = 0.1
  speed_multiplier = 0.8

  [video.OSD]
    display_cycle_speed = true
    display_playback_speed = true
    display_time_remaining = true
    font_size = 40
    margin_left = 10
    margin_top = 10
`

	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf(errFailedToCreateTempFile.Error(), err)
	}

	// Test loading the config
	metadata, err := LoadSessionMetadata(tempFile)
	if err != nil {
		t.Fatalf("LoadSessionMetadata() unexpected error: %v", err)
	}

	// Should use filename as fallback
	expectedTitle := "test_session"
	if metadata.Title != expectedTitle {
		t.Errorf("LoadSessionMetadata() Title = %v, want %v (filename without extension)",
			metadata.Title, expectedTitle)
	}

	if !metadata.IsValid {
		t.Error("LoadSessionMetadata() metadata.IsValid should be true")
	}

}

// TestLoadSessionMetadataWithWhitespaceTitle tests behavior with whitespace-only title
func TestLoadSessionMetadataWithWhitespaceTitle(t *testing.T) {

	// Create a temporary config file with whitespace-only session_title
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "whitespace_test.toml")

	configContent := `
[app]
  session_title = "   "
  logging_level = "info"

[ble]
  sensor_bd_addr = "FA:46:1D:77:C8:E1"
  scan_timeout_secs = 30

[speed]
  speed_threshold = 0.25
  speed_units = "mph"
  smoothing_window = 5
  wheel_circumference_mm = 2155

[video]
  media_player = "mpv"
  file_path = "test_video.mp4"
  window_scale_factor = 1.0
  seek_to_position = "00:00"
  update_interval_secs = 0.1
  speed_multiplier = 0.8

  [video.OSD]
    display_cycle_speed = true
    display_playback_speed = true
    display_time_remaining = true
    font_size = 40
    margin_left = 10
    margin_top = 10
`

	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf(errFailedToCreateTempFile.Error(), err)
	}

	// Test loading the config
	metadata, err := LoadSessionMetadata(tempFile)
	if err != nil {
		t.Fatalf("LoadSessionMetadata() unexpected error: %v", err)
	}

	// Should use filename as fallback when title is only whitespace
	expectedTitle := "whitespace_test"
	if metadata.Title != expectedTitle {
		t.Errorf("LoadSessionMetadata() Title = %v, want %v (filename without extension)",
			metadata.Title, expectedTitle)
	}

}

// TestLoadSessionMetadataValidationErrors tests that validation errors are properly reported
func TestLoadSessionMetadataValidationErrors(t *testing.T) {

	// Create a temporary config file with invalid values
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid_values.toml")

	configContent := `
[app]
  session_title = "Test Session"
  logging_level = "invalid_level"

[ble]
  sensor_bd_addr = "FA:46:1D:77:C8:E1"
  scan_timeout_secs = 30

[speed]
  speed_threshold = 0.25
  speed_units = "mph"
  smoothing_window = 5
  wheel_circumference_mm = 2155

[video]
  media_player = "mpv"
  file_path = "test_video.mp4"
  window_scale_factor = 1.0
  seek_to_position = "00:00"
  update_interval_secs = 0.1
  speed_multiplier = 0.8

  [video.OSD]
    display_cycle_speed = true
    display_playback_speed = true
    display_time_remaining = true
    font_size = 40
    margin_left = 10
    margin_top = 10
`

	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf(errFailedToCreateTempFile.Error(), err)
	}

	// Test loading the invalid config
	metadata, err := LoadSessionMetadata(tempFile)

	// Should return an error
	if err == nil {
		t.Error("LoadSessionMetadata() expected error for invalid log level, got nil")
	}

	// Metadata should still be returned
	if metadata == nil {
		t.Fatal("LoadSessionMetadata() should return metadata even on validation error")
	}

	// Should be marked as invalid
	if metadata.IsValid {
		t.Error("LoadSessionMetadata() metadata.IsValid should be false for validation errors")
	}

	// Should have error message
	if metadata.ErrorMsg == "" {
		t.Error("LoadSessionMetadata() metadata.ErrorMsg should contain validation error")
	}

}
