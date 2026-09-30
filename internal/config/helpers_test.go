package config

import (
	"os"
	"path/filepath"
	"testing"
)

// malformedTOML is file content that cannot be decoded as TOML
const malformedTOML = "[app\nsession_title = "

// writeTestFile writes content to a new file called name in a temporary directory, and returns
// the file path
func writeTestFile(t *testing.T, name, content string) string {

	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create test file %s: %v", name, err)
	}

	return path
}

// newTestVideoFile creates an empty placeholder video file in a temporary directory, and returns
// the file path. Config validation only checks that the video file exists, so no actual video is
// needed here (only the video package needs a playable file)
func newTestVideoFile(t *testing.T) string {

	t.Helper()

	return writeTestFile(t, "test_video.mp4", "")
}

// writeTestConfig saves a valid config (backed by a placeholder video file) to a temporary
// directory, after applying any mutations to it, and returns the config file path
func writeTestConfig(t *testing.T, name string, mutations ...func(*Config)) string {

	t.Helper()

	cfg := NewDefault(newTestVideoFile(t))

	for _, mutate := range mutations {
		mutate(cfg)
	}

	path := filepath.Join(t.TempDir(), name)
	if err := Save(path, cfg, GetVersion()); err != nil {
		t.Fatalf("failed to save test config %s: %v", name, err)
	}

	return path
}

// withMissingVideoFile is a config mutation that points the video file path at a file that does
// not exist
func withMissingVideoFile(cfg *Config) {
	cfg.Video.FilePath = filepath.Join(filepath.Dir(cfg.Video.FilePath), "missing_video.mp4")
}
