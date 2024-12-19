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

// testData holds test constants and configurations
type testData struct {
	filename        string
	windowScale     float64
	updateInterval  float64 // Keep as float64 to match VideoConfig
	speedMultiplier float64
	speedThreshold  float64
	contextTimeout  time.Duration
}

var td = testData{
	filename:        "test.mp4",
	windowScale:     1.0,
	updateInterval:  1.0,
	speedMultiplier: 1.0,
	speedThreshold:  0.1,
	contextTimeout:  5 * time.Second,
}

func init() {
	logger.Initialize("debug")
}

// createTestConfig returns test video and speed configurations
func createTestConfig() (config.VideoConfig, config.SpeedConfig) {
	vc := config.VideoConfig{
		FilePath:          td.filename,
		WindowScaleFactor: td.windowScale,
		UpdateIntervalSec: td.updateInterval,
		SpeedMultiplier:   td.speedMultiplier,
		OnScreenDisplay: config.VideoOSDConfig{
			DisplayPlaybackSpeed: true,
		},
	}

	sc := config.SpeedConfig{
		SpeedThreshold: td.speedThreshold,
	}

	return vc, sc
}

// TestNewPlaybackController verifies controller creation and initialization
func TestNewPlaybackController(t *testing.T) {
	// Create test configuration
	vc, sc := createTestConfig()

	controller, err := NewPlaybackController(vc, sc)
	assert.NotNil(t, controller, "controller should not be nil")
	assert.NoError(t, err, "should create controller without error")
}

// TestPlaybackFlow tests the complete playback flow
func TestPlaybackFlow(t *testing.T) {
	// Create test controller
	controller := createTestController(t)

	// Test configuration
	t.Run("configure player", func(t *testing.T) {
		err := controller.configureMPVPlayer()
		assert.NoError(t, err, "should configure player")
	})

	// Test video loading
	t.Run("load video", func(t *testing.T) {
		err := controller.loadMPVVideo()
		assert.NoError(t, err, "should load video")
	})

	// Test playback control
	t.Run("playback control", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), td.contextTimeout)
		defer cancel()

		speedCtrl := &speed.SpeedController{}
		err := controller.Start(ctx, speedCtrl)
		assert.NoError(t, err, "should start playback")
	})
}

// TestPauseControl tests pause functionality
func TestPauseControl(t *testing.T) {
	// Create test controller
	controller := createTestController(t)

	// Define tests
	tests := []struct {
		name     string
		setPause bool
	}{
		{"pause video", true},
		{"unpause video", false},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.player.SetProperty("pause", mpv.FormatFlag, tt.setPause)
			assert.NoError(t, err, "should set pause state")

			result, err := controller.player.GetProperty("pause", mpv.FormatFlag)
			assert.NoError(t, err, "should get pause state")
			assert.Equal(t, tt.setPause, result.(bool), "pause state should match")
		})
	}
}

// createTestController creates a PlaybackController with default test configurations
func createTestController(t *testing.T) *PlaybackController {
	// Create test configuration
	vc, sc := createTestConfig()

	controller, err := NewPlaybackController(vc, sc)
	assert.NotNil(t, controller, "PlaybackController should not be nil")
	assert.NoError(t, err, "Error while creating PlaybackController")

	return controller
}
