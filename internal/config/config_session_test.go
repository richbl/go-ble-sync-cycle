package config

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

// TestLoadSessionMetadataSuccess tests the successful loading of session metadata
func TestLoadSessionMetadataSuccess(t *testing.T) {

	t.Run("valid config with session title", func(t *testing.T) {
		expectedTitle := "Session Title"

		configFile := writeTestConfig(t, "session.toml", func(cfg *Config) {
			cfg.App.SessionTitle = expectedTitle
		})

		metadata, err := LoadSessionMetadata(configFile)

		if err != nil {
			t.Fatalf("LoadSessionMetadata() returned unexpected error: %v", err)
		}

		if metadata == nil {
			t.Fatal("LoadSessionMetadata() returned nil metadata")

			return
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
			configFile: filepath.Join(t.TempDir(), "non_existent.toml"),
		},
		{
			name:       "malformed config file",
			configFile: writeTestFile(t, "malformed.toml", malformedTOML),
		},
		{
			name:       "config file with missing video file",
			configFile: writeTestConfig(t, "missing_video.toml", withMissingVideoFile),
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

			if metadata != nil {
				t.Error("LoadSessionMetadata() should return nil metadata on error")
			}

		})
	}

}

// TestLoadSessionMetadataMissingVideoFile tests that a missing video file is reported as the
// reason a session config is invalid
func TestLoadSessionMetadataMissingVideoFile(t *testing.T) {

	configFile := writeTestConfig(t, "missing_video.toml", withMissingVideoFile)

	if _, err := LoadSessionMetadata(configFile); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("LoadSessionMetadata() error = %v, want error matching %v", err, fs.ErrNotExist)
	}

}

// TestLoadSessionMetadataTitleFallback tests that the config filename (without extension) is
// used as the session title when session_title is empty or only whitespace
func TestLoadSessionMetadataTitleFallback(t *testing.T) {

	// Define test cases
	tests := []struct {
		name          string
		sessionTitle  string
		filename      string
		expectedTitle string
	}{
		{"empty title", "", "test_session.toml", "test_session"},
		{"whitespace-only title", "   ", "whitespace_test.toml", "whitespace_test"},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			configFile := writeTestConfig(t, tt.filename, func(cfg *Config) {
				cfg.App.SessionTitle = tt.sessionTitle
			})

			metadata, err := LoadSessionMetadata(configFile)
			if err != nil {
				t.Fatalf("LoadSessionMetadata() unexpected error: %v", err)
			}

			if metadata.Title != tt.expectedTitle {
				t.Errorf("LoadSessionMetadata() Title = %v, want %v (filename without extension)",
					metadata.Title, tt.expectedTitle)
			}

			if !metadata.IsValid {
				t.Error("LoadSessionMetadata() metadata.IsValid should be true")
			}

		})
	}

}

// TestLoadSessionMetadataValidationErrors tests that validation errors are properly reported
func TestLoadSessionMetadataValidationErrors(t *testing.T) {

	// Create a config file that is well-formed TOML, but with an invalid log level
	configFile := writeTestConfig(t, "invalid_values.toml", func(cfg *Config) {
		cfg.App.LogLevel = "invalid_level"
	})

	metadata, err := LoadSessionMetadata(configFile)

	// Should fail validation (and not, say, TOML decoding)
	if !errors.Is(err, errInvalidLogLevel) {
		t.Errorf("LoadSessionMetadata() error = %v, want error matching %v", err, errInvalidLogLevel)
	}

	// metadata should be nil
	if metadata != nil {
		t.Error("LoadSessionMetadata() should return nil metadata for validation errors")
	}

}
