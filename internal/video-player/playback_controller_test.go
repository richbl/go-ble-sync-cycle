package video

import (
	"context"
	"testing"
	"time"

	"github.com/gen2brain/go-mpv"
	"github.com/stretchr/testify/assert"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	"github.com/richbl/go-ble-sync-cycle/internal/speed"
)

const (
	testFilename           = "test.mp4"
	defaultWindowScale     = 1.0
	defaultUpdateInterval  = 1
	defaultSpeedMultiplier = 1.0
	speedThreshold         = 0.1
	contextTimeout         = 5 * time.Second
)

// init initializes the logger with the debug level
func init() {
	logger.Initialize("debug")
}

// TestNewPlaybackController verifies PlaybackController creation
func TestNewPlaybackController(t *testing.T) {
	_ = createTestController(t)
}

// createTestController creates a PlaybackController with default test configurations
func createTestController(t *testing.T) *PlaybackController {

	vc := config.VideoConfig{
		FilePath:          testFilename,
		WindowScaleFactor: defaultWindowScale,
		UpdateIntervalSec: defaultUpdateInterval,
		SpeedMultiplier:   defaultSpeedMultiplier,
		OnScreenDisplay: config.VideoOSDConfig{
			DisplayPlaybackSpeed: true,
		},
	}
	sc := config.SpeedConfig{
		SpeedThreshold: speedThreshold,
	}

	controller, err := NewPlaybackController(vc, sc)
	assert.NotNil(t, controller, "PlaybackController should not be nil")
	assert.NoError(t, err, "Error while creating PlaybackController")

	return controller
}

// TestPlaybackControllerStart checks the Start method
func TestPlaybackControllerStart(t *testing.T) {

	controller := createTestController(t)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	speedController := &speed.SpeedController{}
	err := controller.Start(ctx, speedController)
	assert.NoError(t, err, "Error while starting PlaybackController")
}

// TestPlaybackControllerConfigurePlayer tests player configuration
func TestPlaybackControllerConfigurePlayer(t *testing.T) {
	controller := createTestController(t)
	err := controller.configureMPVPlayer()
	assert.NoError(t, err, "Error while configuring MPV player")
}

// TestPlaybackControllerLoadVideoFile checks video file loading
func TestPlaybackControllerLoadVideoFile(t *testing.T) {
	controller := createTestController(t)
	err := controller.loadMPVVideo()
	assert.NoError(t, err, "Error while loading video file")
}

// TestPlaybackControllerSetPauseStatus verifies pause state setting
func TestPlaybackControllerSetPauseStatus(t *testing.T) {

	controller := createTestController(t)

	err := controller.player.SetProperty("pause", mpv.FormatFlag, true)
	assert.NoError(t, err, "Error while setting pause property")

	result, err := controller.player.GetProperty("pause", mpv.FormatFlag)
	assert.NoError(t, err, "Error while getting pause property")
	assert.True(t, result.(bool), "Pause property should be true")
}
