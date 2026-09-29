package config

import (
	"errors"
	"io/fs"
	"testing"
)

// TestNewDefault tests that NewDefault returns a config that passes validation
func TestNewDefault(t *testing.T) {

	videoPath := newTestVideoFile(t)
	cfg := NewDefault(videoPath)

	if cfg.Video.FilePath != videoPath {
		t.Errorf("NewDefault() video file path = %q, want %q", cfg.Video.FilePath, videoPath)
	}

	if !cfg.Video.OnScreenDisplay.ShowOSD {
		t.Error("NewDefault() ShowOSD should be true")
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("NewDefault() config failed validation: %v", err)
	}

}

// TestNewDefaultVideoFileMustExist tests that NewDefault does not create the video file, so
// validation fails until the file exists
func TestNewDefaultVideoFileMustExist(t *testing.T) {

	cfg := NewDefault(newTestVideoFile(t))
	withMissingVideoFile(cfg)

	if err := cfg.Validate(); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Validate() error = %v, want error matching %v", err, fs.ErrNotExist)
	}

}
