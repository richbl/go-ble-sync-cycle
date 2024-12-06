package video

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"

	"github.com/gen2brain/go-mpv"
)

// PlaybackController represents the MPV video player component and its configuration
type PlaybackController struct {
	config      config.VideoConfig
	speedConfig config.SpeedConfig
	player      *mpv.Mpv
}

// NewPlaybackController creates a new video player with the given configuration
func NewPlaybackController(videoConfig config.VideoConfig, speedConfig config.SpeedConfig) (*PlaybackController, error) {
	player := mpv.New()
	if err := player.Initialize(); err != nil {
		return nil, err
	}

	return &PlaybackController{
		config:      videoConfig,
		speedConfig: speedConfig,
		player:      player,
	}, nil
}

// Start configures and starts the MPV video player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.SpeedController) error {

	defer p.player.TerminateDestroy()
	logger.Info("[VIDEO] Starting MPV video player...")

	if err := p.configurePlayer(); err != nil {
		return err
	}

	logger.Info("[VIDEO] Loading video file: " + p.config.FilePath)
	if err := p.loadVideoFile(); err != nil {
		return err
	}

	if err := p.setPauseStatus(false); err != nil {
		logger.Error("[VIDEO] Failed to start video playback: " + err.Error())
	}

	ticker := time.NewTicker(time.Second * time.Duration(p.config.UpdateIntervalSec))
	defer ticker.Stop()

	lastSpeed := 0.0
	logger.Info("[VIDEO] Entering MPV playback loop...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("[VIDEO] Context cancelled. Shutting down video player component")
			return nil
		case <-ticker.C:
			if err := p.updatePlaybackSpeed(speedController, &lastSpeed); err != nil {
				logger.Error("[VIDEO] Error updating playback speed: " + err.Error())
			}
		}

	}

}

// configurePlayer configures the MPV video player
func (p *PlaybackController) configurePlayer() error {

	return p.player.SetOptionString("autofit", strconv.Itoa(int(p.config.WindowScaleFactor*100))+"%")

}

// loadVideoFile loads the video file into the MPV video player
func (p *PlaybackController) loadVideoFile() error {

	return p.player.Command([]string{"loadfile", p.config.FilePath})

}

// updatePlaybackSpeed updates the video playback speed based on the sensor speed
func (p *PlaybackController) updatePlaybackSpeed(speedController *speed.SpeedController, lastSpeed *float64) error {

	currentSpeed := speedController.GetSmoothedSpeed()

	logger.Info("[VIDEO] Sensor speed buffer: [" + strings.Join(speedController.GetSpeedBuffer(), " ") + "]")
	logger.Info("[VIDEO] Smoothed sensor speed: " + strconv.FormatFloat(currentSpeed, 'f', 2, 64))

	if currentSpeed == 0 {
		logger.Info("[VIDEO] No speed detected, so pausing video...")
		return p.setPauseStatus(true)
	}

	// Adjust video playback speed if the current speed is different from the last speed
	if math.Abs(currentSpeed-*lastSpeed) > p.speedConfig.SpeedThreshold {
		newSpeed := (currentSpeed * p.config.SpeedMultiplier) / 10.0
		logger.Info("[VIDEO] Adjusting video speed to " + strconv.FormatFloat(newSpeed, 'f', 2, 64))

		if err := p.setPlaybackSpeed(newSpeed); err != nil {
			return err
		}

	}

	*lastSpeed = currentSpeed
	return nil

}

// setPlaybackSpeed sets the video playback speed
func (p *PlaybackController) setPlaybackSpeed(newSpeed float64) error {

	if err := p.player.SetProperty("speed", mpv.FormatDouble, newSpeed); err != nil {
		logger.Error("[VIDEO] Failed to update video speed: " + err.Error())
		return err
	}

	if p.config.DisplaySpeed {

		if err := p.player.SetOptionString("osd-msg1", fmt.Sprintf("Speed: %.2f", newSpeed)); err != nil {
			return err
		}

	}

	return p.setPauseStatus(false)

}

// setPauseStatus sets the video playback status
func (p *PlaybackController) setPauseStatus(pause bool) error {

	currentPause, err := p.player.GetProperty("pause", mpv.FormatFlag)

	if err != nil {
		logger.Error("[VIDEO] Failed to get pause status: " + err.Error())
		return err
	}

	// Return if the current pause status matches the requested pause status
	if currentPause.(bool) == pause {
		return nil
	}

	// Set the pause status
	if err := p.player.SetProperty("pause", mpv.FormatFlag, pause); err != nil {
		result := "resumed"

		if pause {
			result = "paused"
		}

		logger.Error("[VIDEO] Failed to " + result + " video: " + err.Error())
		return err

	}

	result := "resumed"

	if pause {
		result = "paused"
	}

	logger.Info("[VIDEO] Video " + result + " successfully")
	return nil

}
