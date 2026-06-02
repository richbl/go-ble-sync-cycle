package video

import (
	"testing"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// mpvPlayerFactory creates a new mpvPlayer instance for testing
func mpvPlayerFactory() (*mpvPlayer, error) {
	return newMpvPlayer(logger.BackgroundCtx, config.VideoConfig{})
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
