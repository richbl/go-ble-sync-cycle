// Package video provides video playback control functionality synchronized with speed measurements
package video

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/go-mpv"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

// Common errors for playback control
var (
	ErrOSDUpdate     = errors.New("failed to update OSD")
	ErrPlaybackSpeed = errors.New("failed to set playback speed")
	ErrVideoComplete = errors.New("playback completed: normal exit")
	ErrSpeedUpdate   = errors.New("failed to update video speed")
)

// PlaybackController manages video playback using MPV media player
type PlaybackController struct {
	config      config.VideoConfig
	speedConfig config.SpeedConfig
	player      *mpv.Mpv
}

// NewPlaybackController creates a new video player with the given config
func NewPlaybackController(videoConfig config.VideoConfig, speedConfig config.SpeedConfig) (*PlaybackController, error) {

	player := mpv.New()
	if err := player.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize MPV player: %w", err)
	}

	return &PlaybackController{
		config:      videoConfig,
		speedConfig: speedConfig,
		player:      player,
	}, nil
}

// Start configures and starts the MPV media player, then manages the playback loop and
// synchronizes video speed with the provided speed controller
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.SpeedController) error {

	logger.Info(logger.VIDEO, "starting MPV video player...")
	defer p.player.TerminateDestroy()

	if err := p.configureMPVPlayer(); err != nil {
		return fmt.Errorf("failed to configure MPV player: %w", err)
	}

	logger.Debug(logger.VIDEO, "loading video file:", p.config.FilePath)
	if err := p.loadMPVVideo(); err != nil {
		return fmt.Errorf("failed to load video file: %w", err)
	}

	return p.runPlaybackLoop(ctx, speedController)
}

// runPlaybackLoop runs the main playback loop, updating the video playback speed
func (p *PlaybackController) runPlaybackLoop(ctx context.Context, speedController *speed.SpeedController) error {

	ticker := time.NewTicker(time.Millisecond * time.Duration(p.config.UpdateIntervalSec*1000))
	defer ticker.Stop()

	var lastSpeed float64
	logger.Info(logger.VIDEO, "MPV video playback started")
	logger.Debug(logger.VIDEO, "entering MPV playback loop...")

	for {
		select {
		case <-ctx.Done():
			fmt.Print("\r") // Clear the ^C character
			logger.Info(logger.VIDEO, "user-generated interrupt, stopping MPV video player...")
			return nil
		case <-ticker.C:

			if err := p.handlePlaybackTick(speedController, &lastSpeed); err != nil {

				if errors.Is(err, ErrVideoComplete) {
					return err
				}

				logger.Warn(logger.VIDEO, "playback error:", err.Error())
			}
		}
	}
}

// handlePlaybackTick updates the video playback speed based on the speed controller
func (p *PlaybackController) handlePlaybackTick(speedController *speed.SpeedController, lastSpeed *float64) error {

	// Check for end of file
	reachedEOF, err := p.player.GetProperty("eof-reached", mpv.FormatFlag)
	if err == nil && reachedEOF.(bool) {
		return ErrVideoComplete
	}

	if err := p.updatePlaybackSpeed(speedController, lastSpeed); err != nil {
		if !strings.Contains(err.Error(), "end of file") {
			return fmt.Errorf("error updating playback speed: %w", err)
		}
	}

	return nil
}

// configureMPVPlayer configures the MPV media player
func (p *PlaybackController) configureMPVPlayer() error {

	if err := p.player.SetOptionString("keep-open", "yes"); err != nil {
		return err
	}

	// Set video window size
	if p.config.WindowScaleFactor == 1.0 {
		logger.Debug(logger.VIDEO, "maximizing video window")
		return p.player.SetOptionString("window-maximized", "yes")
	}

	logger.Debug(logger.VIDEO, "scaling video window")
	scalePercent := strconv.Itoa(int(p.config.WindowScaleFactor * 100))

	return p.player.SetOptionString("autofit", scalePercent+"%")
}

// loadMPVVideo loads the video file
func (p *PlaybackController) loadMPVVideo() error {
	return p.player.Command([]string{"loadfile", p.config.FilePath})
}

// updatePlaybackSpeed updates the video playback speed
func (p *PlaybackController) updatePlaybackSpeed(speedController *speed.SpeedController, lastSpeed *float64) error {

	currentSpeed := speedController.GetSmoothedSpeed()
	p.logSpeedInfo(speedController, currentSpeed)

	return p.checkSpeedState(currentSpeed, lastSpeed)
}

// logSpeedInfo logs speed information
func (p *PlaybackController) logSpeedInfo(sc *speed.SpeedController, currentSpeed float64) {

	logger.Debug(logger.VIDEO, "sensor speed buffer: ["+strings.Join(sc.GetSpeedBuffer(), " ")+"]")
	logger.Debug(logger.VIDEO, logger.Magenta+"smoothed sensor speed:",
		strconv.FormatFloat(currentSpeed, 'f', 2, 64), p.speedConfig.SpeedUnits)
}

// checkSpeedState checks the current speed and updates the playback speed
func (p *PlaybackController) checkSpeedState(currentSpeed float64, lastSpeed *float64) error {

	// If no speed detected, pause playback
	if currentSpeed == 0 {
		return p.pausePlayback()
	}

	// If the delta between the current speed and the last speed is greater than the threshold,
	deltaSpeed := math.Abs(currentSpeed - *lastSpeed)
	p.logSpeedDebugInfo(*lastSpeed, deltaSpeed)

	if deltaSpeed > p.speedConfig.SpeedThreshold {
		return p.adjustPlayback(currentSpeed, lastSpeed)
	}

	return nil
}

// logSpeedDebugInfo logs debug information
func (p *PlaybackController) logSpeedDebugInfo(lastSpeed, deltaSpeed float64) {

	logger.Debug(logger.VIDEO, logger.Magenta+"last playback speed:",
		strconv.FormatFloat(lastSpeed, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"sensor speed delta:",
		strconv.FormatFloat(deltaSpeed, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"playback speed update threshold:",
		strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64), p.speedConfig.SpeedUnits)
}

// pausePlayback pauses the video playback
func (p *PlaybackController) pausePlayback() error {

	logger.Debug(logger.VIDEO, "no speed detected, so pausing video")

	if err := p.updateMPVDisplay(0.0, 0.0); err != nil {
		return wrapError(ErrOSDUpdate, err)
	}

	return p.setMPVPauseState(true)
}

// adjustPlayback adjusts the video playback speed
func (p *PlaybackController) adjustPlayback(currentSpeed float64, lastSpeed *float64) error {

	playbackSpeed := (currentSpeed * p.config.SpeedMultiplier) / 10.0
	logger.Debug(logger.VIDEO, logger.Cyan+"updating video playback speed to",
		strconv.FormatFloat(playbackSpeed, 'f', 2, 64))

	if err := p.updateMPVPlaybackSpeed(playbackSpeed); err != nil {
		return wrapError(ErrPlaybackSpeed, err)
	}

	*lastSpeed = currentSpeed

	if err := p.updateMPVDisplay(currentSpeed, playbackSpeed); err != nil {
		return wrapError(ErrOSDUpdate, err)
	}

	return p.setMPVPauseState(false)
}

// updateMPVDisplay updates the MPV OSD
func (p *PlaybackController) updateMPVDisplay(cycleSpeed, playbackSpeed float64) error {

	if !p.config.OnScreenDisplay.ShowOSD {
		return nil
	}

	osdText := p.buildOSDText(cycleSpeed, playbackSpeed)

	return p.player.SetOptionString("osd-msg1", osdText)
}

// buildOSDText builds the MPV OSD text
func (p *PlaybackController) buildOSDText(cycleSpeed, playbackSpeed float64) string {

	// If no speed detected, show "Paused"
	if cycleSpeed == 0 {
		return " Paused"
	}

	var osdText string
	if p.config.OnScreenDisplay.DisplayCycleSpeed {
		osdText += fmt.Sprintf(" Cycle Speed: %.2f %s\n", cycleSpeed, p.speedConfig.SpeedUnits)
	}

	if p.config.OnScreenDisplay.DisplayPlaybackSpeed {
		osdText += fmt.Sprintf(" Playback Speed: %.2fx\n", playbackSpeed)
	}

	return osdText
}

// updateMPVPlaybackSpeed updates the video playback speed
func (p *PlaybackController) updateMPVPlaybackSpeed(playbackSpeed float64) error {
	return p.player.SetProperty("speed", mpv.FormatDouble, playbackSpeed)
}

// setMPVPauseState sets the MPV pause state
func (p *PlaybackController) setMPVPauseState(pause bool) error {
	return p.player.SetProperty("pause", mpv.FormatFlag, pause)
}

// wrapError wraps an error with a specific error type for more context
func wrapError(baseErr error, contextErr error) error {
	return fmt.Errorf("%w: %v", baseErr, contextErr)
}
