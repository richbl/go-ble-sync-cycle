package video

import (
	"context"
	"testing"
	"time"

	"github.com/gen2brain/go-mpv"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"

	"github.com/stretchr/testify/assert"
)

// init initializes the logger with the debug level
func init() {

	// logger needed for testing of video controller component
	logger.Initialize("debug")

}

// TestNewPlaybackController tests the creation of a new PlaybackController
func TestNewPlaybackController(t *testing.T) {

	// Create a video configuration for testing
	videoConfig := config.VideoConfig{
		FilePath:             "test.mp4",
		WindowScaleFactor:    1.0,
		UpdateIntervalSec:    1,
		SpeedMultiplier:      1.0,
		DisplayPlaybackSpeed: true,
	}

	// Create a speed configuration for testing
	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.1,
	}

	// Create a new PlaybackController with the test configurations
	controller, err := NewPlaybackController(videoConfig, speedConfig)
	assert.NotNil(t, controller)
	assert.NoError(t, err)

}

// TestPlaybackController_Start tests the Start method of the PlaybackController
func TestPlaybackController_Start(t *testing.T) {

	// Create a video configuration for testing.
	videoConfig := config.VideoConfig{
		FilePath:             "test.mp4",
		WindowScaleFactor:    1.0,
		UpdateIntervalSec:    1,
		SpeedMultiplier:      1.0,
		DisplayPlaybackSpeed: true,
	}

	// Create a speed configuration for testing
	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.1,
	}

	// Create a new PlaybackController with the test configurations
	controller, err := NewPlaybackController(videoConfig, speedConfig)
	assert.NotNil(t, controller)
	assert.NoError(t, err)

	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a speed controller for testing
	speedController := &speed.SpeedController{}

	// Start the PlaybackController with the test context and speed controller
	err = controller.Start(ctx, speedController)
	assert.NoError(t, err)

}

// TestPlaybackController_configurePlayer tests the configurePlayer method of the PlaybackController
func TestPlaybackController_configurePlayer(t *testing.T) {

	// Create a video configuration for testing
	videoConfig := config.VideoConfig{
		FilePath:             "test.mp4",
		WindowScaleFactor:    1.0,
		UpdateIntervalSec:    1,
		SpeedMultiplier:      1.0,
		DisplayPlaybackSpeed: true,
	}

	// Create a speed configuration for testing
	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.1,
	}

	// Create a new PlaybackController with the test configurations
	controller, err := NewPlaybackController(videoConfig, speedConfig)
	assert.NotNil(t, controller)
	assert.NoError(t, err)

	// Configure the player with the test configuration
	err = controller.configureMPVplayer()
	assert.NoError(t, err)

}

// TestPlaybackController_loadVideoFile tests the loadVideoFile method of the PlaybackController
func TestPlaybackController_loadVideoFile(t *testing.T) {

	// Create a video configuration for testing
	videoConfig := config.VideoConfig{
		FilePath:             "test.mp4",
		WindowScaleFactor:    1.0,
		UpdateIntervalSec:    1,
		SpeedMultiplier:      1.0,
		DisplayPlaybackSpeed: true,
	}

	// Create a speed configuration for testing
	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.1,
	}

	// Create a new PlaybackController with the test configurations
	controller, err := NewPlaybackController(videoConfig, speedConfig)
	assert.NotNil(t, controller)
	assert.NoError(t, err)

	// Load the video file with the test configuration
	err = controller.loadMPVvideo()
	assert.NoError(t, err)

}

// TestPlaybackController_setPauseStatus tests the setPauseStatus method of the PlaybackController
func TestPlaybackController_setPauseStatus(t *testing.T) {

	// Create a video configuration for testing
	videoConfig := config.VideoConfig{
		FilePath:             "test.mp4",
		WindowScaleFactor:    1.0,
		SpeedMultiplier:      1.0,
		UpdateIntervalSec:    1,
		DisplayPlaybackSpeed: true,
	}

	// Create a speed configuration for testing
	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.1,
	}

	// Create a new PlaybackController with the test configurations
	controller, err := NewPlaybackController(videoConfig, speedConfig)
	assert.NotNil(t, controller)
	assert.NoError(t, err)

	// Set the pause status
	err = controller.player.SetProperty("pause", mpv.FormatFlag, true)
	assert.NoError(t, err)

	// check if the pause status is set
	result, err := controller.player.GetProperty("pause", mpv.FormatFlag)
	assert.True(t, result.(bool))
	assert.NoError(t, err)

}
