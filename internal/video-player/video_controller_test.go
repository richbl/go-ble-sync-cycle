package video

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gen2brain/go-mpv"
	"github.com/stretchr/testify/assert"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

// testData holds test constants and configurations
type testData struct {
	filename        string
	windowScale     float64
	SeekToPosition  string
	updateInterval  float64
	speedMultiplier float64
	speedThreshold  float64
	contextTimeout  time.Duration
}

var td = testData{
	filename:        "cycling_test.mp4",
	windowScale:     1.0,
	SeekToPosition:  "0.25",
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
		SeekToPosition:    td.SeekToPosition,
		UpdateIntervalSec: td.updateInterval,
		SpeedMultiplier:   td.speedMultiplier,
		OnScreenDisplay: config.VideoOSDConfig{
			FontSize:             24,
			DisplayCycleSpeed:    true,
			DisplayPlaybackSpeed: true,
			DisplayTimeRemaining: true,
			ShowOSD:              true,
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
		err := controller.configurePlayback()
		assert.NoError(t, err, "should configure player")
	})

	// Test video loading
	t.Run("load video", func(t *testing.T) {

		err := controller.mpvPlayer.Command([]string{"loadfile", controller.videoConfig.FilePath})
		assert.NoError(t, err, "should load video")
	})

	// Test playback control
	t.Run("playback control", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), td.contextTimeout)
		defer cancel()

		speedCtrl := &speed.Controller{}
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

			err := controller.mpvPlayer.SetProperty("pause", mpv.FormatFlag, tt.setPause)
			assert.NoError(t, err, "should set pause state")

			result, err := controller.mpvPlayer.GetProperty("pause", mpv.FormatFlag)
			assert.NoError(t, err, "should get pause state")

			// Check if result is a boolean
			propertyPause, ok := result.(bool)
			if !ok {
				t.Error(fmt.Errorf(errTypeFormat, errUnsupportedType, propertyPause))
			}

			assert.Equal(t, tt.setPause, propertyPause, "pause state should match")
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

// TestConfigurePlaybackWindow tests different window configuration scenarios
func TestConfigurePlaybackWindow(t *testing.T) {

	controller := createTestController(t)

	// Test full screen scenario
	controller.videoConfig.WindowScaleFactor = 1.0
	err := controller.configurePlaybackWindow()
	assert.NoError(t, err, "should configure full screen playback window")

	// Test scaled window scenario
	controller.videoConfig.WindowScaleFactor = 0.5
	err = controller.configurePlaybackWindow()
	assert.NoError(t, err, "should configure scaled playback window")
}

// TestConfigureKeepOpen tests that the player is configured to keep open after playback
func TestConfigureKeepOpen(t *testing.T) {

	controller := createTestController(t)

	err := controller.configureKeepOpen()
	assert.NoError(t, err, "should configure player to keep open after playback")
}

// TestConfigureOSD checks OSD configuration
func TestConfigureOSD(t *testing.T) {

	controller := createTestController(t)

	// Test OSD enabled
	controller.osdConfig.ShowOSD = true
	err := controller.configureOSD()
	assert.NoError(t, err, "should configure OSD when enabled")

	// Test OSD disabled
	controller.osdConfig.ShowOSD = false
	err = controller.configureOSD()
	assert.NoError(t, err, "should not configure OSD when disabled")
}

// TestSeekToStartPosition checks seeking to a start position
func TestSeekToStartPosition(t *testing.T) {

	controller := createTestController(t)

	err := controller.seekToStartPosition()
	assert.NoError(t, err, "should seek to start position")
}

// TestHandleZeroSpeed tests the zero speed handling logic
func TestHandleZeroSpeed(t *testing.T) {

	controller := createTestController(t)

	err := controller.handleZeroSpeed()
	assert.NoError(t, err, "should handle zero speed without error")
}
