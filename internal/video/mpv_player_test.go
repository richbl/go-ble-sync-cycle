package video

import (
	"testing"
)

// mpvPlayerFactory creates a new mpvPlayer instance for testing
func mpvPlayerFactory() (mediaPlayer, error) {
	return newMpvPlayer()
}

func TestMpvPlayerLifecycle(t *testing.T) {
	testPlayerLifecycle(t, mpvPlayerFactory)
}

func TestMpvPlayerPlaybackControls(t *testing.T) {
	testPlayerPlaybackControls(t, mpvPlayerFactory)
}

func TestMpvPlayerConfiguration(t *testing.T) {
	testPlayerConfiguration(t, mpvPlayerFactory, "MPV")
}
