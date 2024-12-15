package video

import (
	"context"
	"errors"
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

// Start configures and starts the MPV media player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.SpeedController) error {

	logger.Info(logger.VIDEO, "Starting MPV video player...")

	// Defer terminating and destroying the MPV media player
	defer p.player.TerminateDestroy()

	// Configure the MPV media player
	if err := p.configureMPVPlayer(); err != nil {
		return err
	}

	// Load the video file into MPV
	logger.Debug(logger.VIDEO, "Loading video file: "+p.config.FilePath)
	if err := p.loadMPVvideo(); err != nil {
		return err
	}

	// Set the MPV playback loop interval
	ticker := time.NewTicker(time.Millisecond * time.Duration(p.config.UpdateIntervalSec*1000))
	defer ticker.Stop()

	lastSpeed := 0.0
	logger.Debug(logger.VIDEO, "Entering MPV playback loop...")

	for {
		select {
		case <-ctx.Done():
			logger.Info(logger.VIDEO, "Context cancelled. Shutting down video player component")
			return nil
		case <-ticker.C:
			if err := p.updatePlaybackSpeed(speedController, &lastSpeed); err != nil {
				logger.Warn(logger.VIDEO, "Error updating playback speed: "+err.Error())
			}
		}
	}
}

// configureMPVPlayer configures the MPV video player
func (p *PlaybackController) configureMPVPlayer() error {

	// If the window scale factor is set to 1.0, set the window to be maximized
	if p.config.WindowScaleFactor == 1.0 {
		logger.Debug(logger.VIDEO, "Maximizing video window")
		return p.player.SetOptionString("window-maximized", "yes")
	}

	return p.player.SetOptionString("autofit", strconv.Itoa(int(p.config.WindowScaleFactor*100))+"%")
}

// loadMPVvideo loads the video file into the MPV video player
func (p *PlaybackController) loadMPVvideo() error {
	return p.player.Command([]string{"loadfile", p.config.FilePath})
}

// updatePlaybackSpeed updates the video playback speed based on the sensor speed
func (p *PlaybackController) updatePlaybackSpeed(speedController *speed.SpeedController, lastSpeed *float64) error {

	// Get and display (log) the smoothed sensor speed
	currentSpeed := speedController.GetSmoothedSpeed()
	p.logSpeedInfo(speedController, currentSpeed)

	// Check current sensor speed and adjust video playback speed if required
	if err := p.checkSpeedState(currentSpeed, lastSpeed); err != nil {
		return err
	}

	return nil
}

// logSpeedInfo logs the sensor speed details
func (p *PlaybackController) logSpeedInfo(sc *speed.SpeedController, currentSpeed float64) {
	logger.Debug(logger.VIDEO, "Sensor speed buffer: ["+strings.Join(sc.GetSpeedBuffer(), " ")+"]")
	logger.Info(logger.VIDEO, logger.Magenta+"Smoothed sensor speed: "+strconv.FormatFloat(currentSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
}

// checkSpeedState checks the current sensor speed and adjusts video playback
func (p *PlaybackController) checkSpeedState(currentSpeed float64, lastSpeed *float64) error {

	// Pause the video playback if no speed is detected
	if currentSpeed == 0 {
		return p.pausePlayback()
	}

	// Calculate the delta between the current and last sensor speed
	deltaSpeed := math.Abs(currentSpeed - *lastSpeed)

	logger.Debug(logger.VIDEO, logger.Magenta+"Last playback speed: "+strconv.FormatFloat(*lastSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"Sensor speed delta: "+strconv.FormatFloat(deltaSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"Playback speed update threshold: "+strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)

	// Adjust the video playback speed if the sensor speed has changed beyond threshold value
	if deltaSpeed > p.speedConfig.SpeedThreshold {
		return p.adjustPlayback(currentSpeed, lastSpeed)
	}

	return nil
}

// pausePlayback pauses the video playback in the MPV media player
func (p *PlaybackController) pausePlayback() error {

	logger.Debug(logger.VIDEO, "No speed detected, so pausing video")

	// Update the on-screen display
	if err := p.updateMPVdisplay(0.0, 0.0); err != nil {
		return errors.New("failed to update OSD: " + err.Error())
	}

	// Pause the video
	return p.setMPVpauseState(true)
}

// adjustPlayback adjusts the video playback speed
func (p *PlaybackController) adjustPlayback(currentSpeed float64, lastSpeed *float64) error {

	// Calculate the new playback speed
	playbackSpeed := (currentSpeed * p.config.SpeedMultiplier) / 10.0
	logger.Info(logger.VIDEO, logger.Cyan+"Updating video playback speed to "+strconv.FormatFloat(playbackSpeed, 'f', 2, 64))

	// Update the video playback speed
	if err := p.updateMPVplaybackSpeed(playbackSpeed); err != nil {
		return errors.New("failed to set playback speed: " + err.Error())
	}

	// Update the last sensor speed value sent to the media player to the current value
	*lastSpeed = currentSpeed

	// Update the on-screen display
	if err := p.updateMPVdisplay(currentSpeed, playbackSpeed); err != nil {
		return errors.New("failed to update OSD: " + err.Error())
	}

	// Unpause the video
	return p.setMPVpauseState(false)
}

// updateMPVdisplay updates the MPV media player on-screen display
func (p *PlaybackController) updateMPVdisplay(cycleSpeed, playbackSpeed float64) error {

	// Return if no OSD options are enabled in TOML
	if !p.config.OnScreenDisplay.ShowOSD {
		return nil
	}

	// Build the OSD message based on TOML configuration
	var osdMsg string
	if p.config.OnScreenDisplay.DisplayCycleSpeed {
		osdMsg += "Sensor Speed: " + strconv.FormatFloat(cycleSpeed, 'f', 2, 64) + " " + p.speedConfig.SpeedUnits + "\n"
	}

	if p.config.OnScreenDisplay.DisplayPlaybackSpeed {
		osdMsg += "Playback Speed: " + strconv.FormatFloat(playbackSpeed, 'f', 2, 64) + "\n"
	}

	// Update the MPV media player on-screen display (OSD)
	return p.player.SetOptionString("osd-msg1", osdMsg)
}

// updateMPVplaybackSpeed sets the video playback speed
func (p *PlaybackController) updateMPVplaybackSpeed(playbackSpeed float64) error {

	// Set the new playback speed in MPV
	if err := p.player.SetProperty("speed", mpv.FormatDouble, playbackSpeed); err != nil {
		return errors.New("failed to update video speed: " + err.Error())
	}

	return nil
}

// setMPVpauseState sets the video playback pause state
func (p *PlaybackController) setMPVpauseState(pause bool) error {

	// Get the current pause state from MPV
	currentPause, err := p.player.GetProperty("pause", mpv.FormatFlag)
	if err != nil {
		return err
	}

	// Return if the current pause state matches the requested pause state
	if pauseState, ok := currentPause.(bool); ok && pauseState == pause {
		return nil
	}

	// Set the new pause state in MPV
	if err := p.player.SetProperty("pause", mpv.FormatFlag, pause); err != nil {
		return err
	}

	pauseState := "resumed"
	if pause {
		pauseState = "paused"
	}

	logger.Debug(logger.VIDEO, "Video "+pauseState+" successfully")
	return nil
}
