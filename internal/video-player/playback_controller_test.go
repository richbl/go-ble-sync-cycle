package video

import (
	"context"
	"testing"
	"time"

	"github.com/gen2brain/go-mpv"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	"github.com/richbl/go-ble-sync-cycle/internal/speed"

	"github.com/stretchr/testify/assert"
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

// TestNewPlaybackController tests the creation of a new PlaybackController
func TestNewPlaybackController(t *testing.T) {
	_ = createTestController(t)
}

// createTestController creates a PlaybackController with default test configurations
func createTestController(t *testing.T) *PlaybackController {

	// Create test Video and Speed configurations
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

	// Create new PlaybackController
	controller, err := NewPlaybackController(vc, sc)
	assert.NotNil(t, controller, "PlaybackController should not be nil")
	assert.NoError(t, err, "Error while creating PlaybackController")

	return controller
}

// TestPlaybackControllerStart tests the Start method of the PlaybackController
func TestPlaybackControllerStart(t *testing.T) {

	// Create test controller
	controller := createTestController(t)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	// Create new SpeedController
	speedController := &speed.SpeedController{}
	err := controller.Start(ctx, speedController)
	assert.NoError(t, err, "Error while starting PlaybackController")
}

// TestPlaybackControllerConfigurePlayer tests the configurePlayer method of the PlaybackController
func TestPlaybackControllerConfigurePlayer(t *testing.T) {

	// Create test controller
	controller := createTestController(t)
	err := controller.configureMPVPlayer()
	assert.NoError(t, err, "Error while configuring MPV player")
}

// TestPlaybackControllerLoadVideoFile tests the loadVideoFile method of the PlaybackController
func TestPlaybackControllerLoadVideoFile(t *testing.T) {

	// Create test controller
	controller := createTestController(t)
	err := controller.loadMPVvideo()
	assert.NoError(t, err, "Error while loading video file")
}

// TestPlaybackControllerSetPauseStatus tests the setPauseStatus method of the PlaybackController
func TestPlaybackControllerSetPauseStatus(t *testing.T) {

	// Create test controller
	controller := createTestController(t)

	// Set and check pause status
	err := controller.player.SetProperty("pause", mpv.FormatFlag, true)
	assert.NoError(t, err, "Error while setting pause property")

	result, err := controller.player.GetProperty("pause", mpv.FormatFlag)
	assert.NoError(t, err, "Error while getting pause property")
	assert.True(t, result.(bool), "Pause property should be true")
}
