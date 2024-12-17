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
	ErrOSDUpdate       = errors.New("failed to update OSD")
	ErrPlaybackSpeed   = errors.New("failed to set playback speed")
	ErrVideoComplete   = errors.New("playback completed: normal exit")
	ErrSpeedUpdate     = errors.New("failed to update video speed")
)

// wrapError wraps an error with a specific error type for more context
func wrapError(baseErr error, contextErr error) error {
	return fmt.Errorf("%w: %v", baseErr, contextErr)
}

// PlaybackController manages video playback using MPV media player
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

	logger.Info(logger.VIDEO, "starting MPV video player...")
	defer p.player.TerminateDestroy()

	if err := p.configureMPVPlayer(); err != nil {
		return err
	}

	logger.Debug(logger.VIDEO, "loading video file: "+p.config.FilePath)
	if err := p.loadMPVVideo(); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Millisecond * time.Duration(p.config.UpdateIntervalSec*1000))
	defer ticker.Stop()

	var lastSpeed float64
	logger.Debug(logger.VIDEO, "entering MPV playback loop...")

	for {
		select {
		case <-ctx.Done():
			logger.Info(logger.VIDEO, "context cancelled, stopping video player...")
			return nil
		case <-ticker.C:
			reachedEOF, err := p.player.GetProperty("eof-reached", mpv.FormatFlag)
			if err == nil && reachedEOF.(bool) {
				return ErrVideoComplete
			}

			if err := p.updatePlaybackSpeed(speedController, &lastSpeed); err != nil {
				if !strings.Contains(err.Error(), "end of file") {
					logger.Warn(logger.VIDEO, "error updating playback speed: "+err.Error())
				}
			}
		}
	}
}

// configureMPVPlayer configures the MPV video player settings
func (p *PlaybackController) configureMPVPlayer() error {

	if p.config.WindowScaleFactor == 1.0 {
		logger.Debug(logger.VIDEO, "maximizing video window")
		return p.player.SetOptionString("window-maximized", "yes")
	}

	if err := p.player.SetOptionString("keep-open", "yes"); err != nil {
		return err
	}

	return p.player.SetOptionString("autofit", strconv.Itoa(int(p.config.WindowScaleFactor*100))+"%")
}

// loadMPVVideo loads the video file into the MPV video player
func (p *PlaybackController) loadMPVVideo() error {
	return p.player.Command([]string{"loadfile", p.config.FilePath})
}

// updatePlaybackSpeed updates the video playback speed based on the sensor speed
func (p *PlaybackController) updatePlaybackSpeed(speedController *speed.SpeedController, lastSpeed *float64) error {

	currentSpeed := speedController.GetSmoothedSpeed()
	p.logSpeedInfo(speedController, currentSpeed)

	return p.checkSpeedState(currentSpeed, lastSpeed)
}

// logSpeedInfo logs the sensor speed details
func (p *PlaybackController) logSpeedInfo(sc *speed.SpeedController, currentSpeed float64) {
	logger.Debug(logger.VIDEO, "sensor speed buffer: ["+strings.Join(sc.GetSpeedBuffer(), " ")+"]")
	logger.Info(logger.VIDEO, logger.Magenta+"smoothed sensor speed: "+strconv.FormatFloat(currentSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
}

// checkSpeedState checks the current sensor speed and adjusts video playback
func (p *PlaybackController) checkSpeedState(currentSpeed float64, lastSpeed *float64) error {

	if currentSpeed == 0 {
		return p.pausePlayback()
	}

	deltaSpeed := math.Abs(currentSpeed - *lastSpeed)

	logger.Debug(logger.VIDEO, logger.Magenta+"last playback speed: "+strconv.FormatFloat(*lastSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"sensor speed delta: "+strconv.FormatFloat(deltaSpeed, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"playback speed update threshold: "+strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64)+" "+p.speedConfig.SpeedUnits)

	if deltaSpeed > p.speedConfig.SpeedThreshold {
		return p.adjustPlayback(currentSpeed, lastSpeed)
	}

	return nil
}

// pausePlayback pauses the video playback in the MPV media player
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
	logger.Info(logger.VIDEO, logger.Cyan+"updating video playback speed to "+strconv.FormatFloat(playbackSpeed, 'f', 2, 64))

	if err := p.updateMPVPlaybackSpeed(playbackSpeed); err != nil {
		return wrapError(ErrPlaybackSpeed, err)
	}

	*lastSpeed = currentSpeed

	if err := p.updateMPVDisplay(currentSpeed, playbackSpeed); err != nil {
		return wrapError(ErrOSDUpdate, err)
	}

	return p.setMPVPauseState(false)
}

// updateMPVDisplay updates the MPV media player on-screen display
func (p *PlaybackController) updateMPVDisplay(cycleSpeed, playbackSpeed float64) error {

	if !p.config.OnScreenDisplay.ShowOSD {
		return nil
	}

	var osdText string
	if cycleSpeed > 0 {
		osdText = fmt.Sprintf("Cycle Speed: %.2f %s\nPlayback Speed: %.2fx", 
			cycleSpeed, p.speedConfig.SpeedUnits, playbackSpeed)
	} else {
		osdText = "Paused"
	}

	return p.player.SetProperty("osd-msg", mpv.FormatString, osdText)
}

// updateMPVPlaybackSpeed sets the video playback speed
func (p *PlaybackController) updateMPVPlaybackSpeed(playbackSpeed float64) error {
	return p.player.SetProperty("speed", mpv.FormatDouble, playbackSpeed)
}

// setMPVPauseState sets the video playback pause state
func (p *PlaybackController) setMPVPauseState(pause bool) error {
	return p.player.SetProperty("pause", mpv.FormatFlag, pause)
}
