package video

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupMpvTest initializes a player with a test video file.
// It returns the player and a cleanup function.
// If the player cannot be initialized (e.g., mpv is not installed), it skips the test.
func setupMpvTest(t *testing.T) (*mpvPlayer, func()) {
	t.Helper()

	// Use the existing test video file.
	filePath, err := filepath.Abs("cycling_test.mp4")
	if err != nil {
		t.Fatalf("Failed to get absolute path for test video: %v", err)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Test video file not found at %s", filePath)
	}

	player, err := newMpvPlayer()
	if err != nil {
		t.Skipf("Skipping MPV tests: failed to create player (is mpv installed?): %v", err)
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

func TestMpvPlayerLifecycle(t *testing.T) {
	_, cleanup := setupMpvTest(t)
	defer cleanup()

	// The setup function already loads the file, so this test primarily checks initialization and cleanup
}

func TestMpvPlayerPlaybackControls(t *testing.T) {
	player, cleanup := setupMpvTest(t)
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

func TestMpvPlayerConfiguration(t *testing.T) {
	player, cleanup := setupMpvTest(t)
	defer cleanup()

	t.Run("setFullscreen", func(t *testing.T) {
		if err := player.setFullscreen(true); err != nil {
			t.Errorf("setFullscreen(true) error = %v", err)
		}
	})

	t.Run("setKeepOpen", func(t *testing.T) {
		if err := player.setKeepOpen(true); err != nil {
			t.Errorf("setKeepOpen(true) error = %v", err)
		}
	})

	t.Run("setOSD", func(t *testing.T) {
		opts := osdConfig{fontSize: 50, marginX: 20, marginY: 20}
		if err := player.setOSD(opts); err != nil {
			t.Errorf("setOSD() error = %v", err)
		}
	})

	t.Run("showOSDText", func(t *testing.T) {
		if err := player.showOSDText("Hello MPV"); err != nil {
			t.Errorf("showOSDText() error = %v", err)
		}
	})
}
