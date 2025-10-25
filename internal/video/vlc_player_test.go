package video

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupVlcTest initializes a player with a test video file.
// It returns the player and a cleanup function.
// If the player cannot be initialized (e.g., vlc is not installed), it skips the test.
func setupVlcTest(t *testing.T) (*vlcPlayer, func()) {
	t.Helper()

	// Use the existing test video file.
	filePath, err := filepath.Abs("cycling_test.mp4")
	if err != nil {
		t.Fatalf("Failed to get absolute path for test video: %v", err)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Test video file not found at %s", filePath)
	}

	player, err := newVLCPlayer()
	if err != nil {
		t.Skipf("Skipping VLC tests: failed to create player (is vlc installed?): %v", err)
	}

	if err := player.loadFile(filePath); err != nil {
		t.Fatalf("loadFile() failed: %v", err)
	}

	// Allow some time for the file to load before setting properties
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		player.terminatePlayer()
	}

	return player, cleanup
}

func TestVLCPlayerLifecycle(t *testing.T) {
	_, cleanup := setupVlcTest(t)
	defer cleanup()

	// The setup function already loads the file, so this test primarily checks initialization and cleanup.
}

func TestVLCPlayerPlaybackControls(t *testing.T) {
	player, cleanup := setupVlcTest(t)
	defer cleanup()

	t.Run("setSpeed", func(t *testing.T) {
		if err := player.setSpeed(1.5); err != nil {
			t.Errorf("setSpeed() error = %v", err)
		}
	})

	t.Run("setPause", func(t *testing.T) {
		if err := player.setPause(true); err != nil {
			t.Errorf("setPause(true) error = %v", err)
		}
		if err := player.setPause(false); err != nil {
			t.Errorf("setPause(false) error = %v", err)
		}
	})

	t.Run("seek", func(t *testing.T) {
		if err := player.seek("10"); err != nil {
			t.Errorf("seek() error = %v", err)
		}
	})
}

func TestVLCPlayerConfiguration(t *testing.T) {
	player, cleanup := setupVlcTest(t)
	defer cleanup()

	t.Run("setFullscreen", func(t *testing.T) {
		if err := player.setFullscreen(true); err != nil {
			t.Errorf("setFullscreen(true) error = %v", err)
		}
	})

	t.Run("setKeepOpen", func(t *testing.T) {
		// This is a no-op for VLC, so we just check for no error.
		if err := player.setKeepOpen(true); err != nil {
			t.Errorf("setKeepOpen(true) should not return an error, but got %v", err)
		}
	})

	t.Run("setOSD", func(t *testing.T) {
		opts := osdConfig{fontSize: 50, marginX: 20, marginY: 20}
		if err := player.setOSD(opts); err != nil {
			t.Errorf("setOSD() error = %v", err)
		}
	})

	t.Run("showOSDText", func(t *testing.T) {
		if err := player.showOSDText("Hello VLC"); err != nil {
			t.Errorf("showOSDText() error = %v", err)
		}
	})
}

func TestVLCPlayerparseTimePosition(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{"MM:SS format", "01:30", 90000, false},
		{"SS format", "45", 45000, false},
		{"Zero seconds", "0", 0, false},
		{"Invalid MM:SS", "01:60", 0, true},
		{"Invalid format", "abc", 0, true},
		{"Negative seconds", "-10", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTimePosition(tc.input)

			if (err != nil) != tc.hasError {
				t.Fatalf("parseTimePosition() error = %v, wantErr %v", err, tc.hasError)
			}

			if !tc.hasError && got != tc.expected {
				t.Errorf("parseTimePosition() = %v, want %v", got, tc.expected)
			}
		})
	}
}
