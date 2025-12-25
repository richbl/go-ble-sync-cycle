package video

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// playerFactory is a function type that creates a media player instance
type playerFactory func() (mediaPlayer, error)

// testVideoPath returns the absolute path to the test video file
func testVideoPath(t *testing.T) string {

	t.Helper()

	filePath, err := filepath.Abs("test_video.mp4")
	if err != nil {
		t.Fatalf("Failed to get absolute path for test video: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Test video file not found at %s", filePath)
	}

	return filePath
}

// setupPlayerTest is a generic setup function for any mediaPlayer implementation
//
//nolint:ireturn // Generic function returning T
func setupPlayerTest[T mediaPlayer](t *testing.T, factory func() (T, error)) (T, func()) {

	t.Helper()
	player, err := factory()
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	filePath := testVideoPath(t)

	if err := player.loadFile(filePath); err != nil {
		player.terminatePlayer()
		t.Fatalf("loadFile(%s) failed: %v", filePath, err)
	}

	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		player.terminatePlayer()
	}

	return player, cleanup
}

// testPlayerLifecycle tests basic player initialization and cleanup
func testPlayerLifecycle(t *testing.T, factory playerFactory) {

	t.Helper()

	// The setup function already loads the file, so this test primarily checks
	// initialization and cleanup
	_, cleanup := setupPlayerTest(t, factory)
	defer cleanup()

}

// testPlayerPlaybackControls tests playback control methods
func testPlayerPlaybackControls(t *testing.T, factory playerFactory) {

	t.Helper()

	player, cleanup := setupPlayerTest(t, factory)
	defer cleanup()

	t.Run(setSpeed, func(t *testing.T) {

		if err := player.setSpeed(1.5); err != nil {
			t.Errorf("setSpeed() error = %v", err)
		}

	})

	t.Run(setPause, func(t *testing.T) {

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

// testPlayerConfiguration tests configuration methods
func testPlayerConfiguration(t *testing.T, factory playerFactory, playerName string) {

	t.Helper()

	player, cleanup := setupPlayerTest(t, factory)
	defer cleanup()

	t.Run("setFullscreen", func(t *testing.T) {

		if err := player.setFullscreen(true); err != nil {
			t.Errorf("setFullscreen(true) error = %v", err)
		}

	})

	t.Run("setKeepOpen", func(t *testing.T) {
		err := player.setKeepOpen(true)

		// VLC returns nil (no-op), MPV should also return nil on success
		if err != nil {
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

		if err := player.showOSDText("Hello " + playerName); err != nil {
			t.Errorf("showOSDText() error = %v", err)
		}

	})
}
