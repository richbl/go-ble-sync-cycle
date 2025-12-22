package video

import (
	"testing"
)

// mpvPlayerFactory creates a new mpvPlayer instance for testing
func mpvPlayerFactory() (*mpvPlayer, error) {
	return newMpvPlayer()
}

func TestMpvPlayerLifecycle(t *testing.T) {
	testPlayerLifecycle(t, func() (mediaPlayer, error) { return mpvPlayerFactory() })
}

func TestMpvPlayerPlaybackControls(t *testing.T) {
	testPlayerPlaybackControls(t, func() (mediaPlayer, error) { return mpvPlayerFactory() })
}

func TestMpvPlayerConfiguration(t *testing.T) {
	testPlayerConfiguration(t, func() (mediaPlayer, error) { return mpvPlayerFactory() }, "MPV")
}
