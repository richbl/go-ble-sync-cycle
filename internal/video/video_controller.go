package video

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

// PlaybackController manages video playback
type PlaybackController struct {

	// Configuration
	videoConfig config.VideoConfig
	speedConfig config.SpeedConfig
	osdConfig   osdConfig

	// Media player state
	player              mediaPlayer
	speedState          *speedState
	speedUnitMultiplier float64
}

// speedState holds the state of the speedController speed
type speedState struct {
	current float64
	last    float64
}

const (
	// Divisor used to convert speed relative to playback rate
	// e.g., a speed of 10 mph = 1.0x video playback (hence divisor of 10)
	speedDivisor = 10.0
)

// speedUnitConversion maps units of speed (mph, km/h) to their multiplier for consistent playback speed
var speedUnitConversion = map[string]float64{
	config.SpeedUnitsKMH: 1.60934,
	config.SpeedUnitsMPH: 1.0,
}

// NewPlaybackController creates a new video player instance with the given config
func NewPlaybackController(videoConfig config.VideoConfig, speedConfig config.SpeedConfig) (*PlaybackController, error) {

	var player mediaPlayer
	var err error

	switch videoConfig.MediaPlayer {

	case config.MediaPlayerMPV:
		player, err = newMpvPlayer()

	case config.MediaPlayerVLC:
		player, err = newVLCPlayer()

	default:
		return nil, fmt.Errorf(errFormat, "unsupported media player", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s player: %w", videoConfig.MediaPlayer, err)
	}

	return &PlaybackController{
		videoConfig: videoConfig,
		speedConfig: speedConfig,
		osdConfig:   newOSDConfig(videoConfig.OnScreenDisplay),
		player:      player,
		speedState:  &speedState{},
	}, nil
}

// newOSDConfig creates a new OSD configuration from the video config
func newOSDConfig(displayConfig config.VideoOSDConfig) osdConfig {
	return osdConfig{
		showOSD:              displayConfig.ShowOSD,
		fontSize:             displayConfig.FontSize,
		displayCycleSpeed:    displayConfig.DisplayCycleSpeed,
		displayPlaybackSpeed: displayConfig.DisplayPlaybackSpeed,
		displayTimeRemaining: displayConfig.DisplayTimeRemaining,
		marginX:              displayConfig.MarginX,
		marginY:              displayConfig.MarginY,
	}
}

// Start configures and starts playback of the media player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.Controller) error {

	logger.Info(logger.VIDEO, "starting", p.videoConfig.MediaPlayer, "video playback...")

	defer p.player.terminatePlayer()

	// Configure the media player
	if err := p.configurePlayback(); err != nil {
		return fmt.Errorf("failed to configure %s video playback: %w", p.videoConfig.MediaPlayer, err)
	}

	// Start the event callback loop for the media player
	if err := p.eventLoop(ctx, speedController); err != nil {
		return err
	}

	return nil
}

// configurePlayback sets up the player window based on configuration
func (p *PlaybackController) configurePlayback() error {

	if err := p.player.loadFile(p.videoConfig.FilePath); err != nil {
		return fmt.Errorf("failed to load video file: %s :%w", p.videoConfig.FilePath, err)
	}

	if err := p.player.setupEvents(); err != nil {
		return fmt.Errorf(errFormat, "failed to setup player events", err)
	}

	if err := p.player.setFullscreen(p.videoConfig.WindowScaleFactor == 1.0); err != nil {
		return err
	}

	if err := p.player.setKeepOpen(true); err != nil {
		return err
	}

	if p.osdConfig.showOSD {
		if err := p.player.setOSD(p.osdConfig); err != nil {
			return err
		}
	}

	if err := p.player.seek(p.videoConfig.SeekToPosition); err != nil {
		return err
	}

	// Precalculate playback speed multiplier based on speed units
	p.speedUnitMultiplier = p.videoConfig.SpeedMultiplier / (speedUnitConversion[p.speedConfig.SpeedUnits] * speedDivisor)

	return nil
}

// eventLoop is the main event loop for the media player
func (p *PlaybackController) eventLoop(ctx context.Context, speedController *speed.Controller) error {

	// Start a ticker to check updates from SpeedController

	ticker := time.NewTicker(time.Duration(p.videoConfig.UpdateIntervalSec * float64(time.Second)))
	defer ticker.Stop()

	for {
		// Check player events (give priority to video completion)
		if err := p.handlePlayerEvents(); err != nil {
			return err
		}

		select {
		case <-ticker.C:

			if err := p.updateSpeedFromController(speedController); err != nil {
				logger.Warn(logger.VIDEO, "speed update error:", err)
			}

		case <-ctx.Done():

			fmt.Print("\r") // Clear the ^C character from the terminal line
			logger.Info(logger.VIDEO, "interrupt detected, stopping", p.videoConfig.MediaPlayer, "video playback...")

			return nil // No need to show this context cancellation error
		}
	}

}

// handlePlayerEvents handles callback events from the media player
func (p *PlaybackController) handlePlayerEvents() error {

	event := p.player.waitEvent(0) // Use a non-blocking wait
	if event != nil && event.id == eventEndFile {
		return fmt.Errorf("%w", errVideoComplete)
	}

	return nil
}

// updateSpeedFromController manages updates from the speedController component
func (p *PlaybackController) updateSpeedFromController(speedController *speed.Controller) error {

	p.speedState.current = speedController.GetSmoothedSpeed()
	p.logDebugInfo(speedController)

	if p.speedState.current == 0 {
		return p.handleZeroSpeed()
	}

	if p.shouldUpdateSpeed() {
		return p.updateSpeed()
	}

	return nil
}

// handleZeroSpeed handles the case when no speed is detected
func (p *PlaybackController) handleZeroSpeed() error {

	logger.Debug(logger.VIDEO, "no speed detected, pausing video")

	if err := p.updateDisplay(0.0, 0.0); err != nil {
		return fmt.Errorf(errFormat, errOSDUpdate, err)
	}

	return p.player.setPause(true)
}

// shouldUpdateSpeed determines if the playback speed needs updating
func (p *PlaybackController) shouldUpdateSpeed() bool {

	// Always update the speed if "display time remaining" option is enabled
	// Else update only if the speed delta is greater than the configured speed threshold
	return p.osdConfig.displayTimeRemaining ||
		(math.Abs(p.speedState.current-p.speedState.last) > p.speedConfig.SpeedThreshold)
}

// updateSpeed adjusts the playback speed based on current speed
func (p *PlaybackController) updateSpeed() error {

	// Update the playback speed based on current speed and unit multiplier
	playbackSpeed := p.speedState.current * p.speedUnitMultiplier

	logger.Debug(logger.VIDEO, logger.Cyan+"updating video playback speed to",
		strconv.FormatFloat(playbackSpeed, 'f', 2, 64)+"x")

	if err := p.player.setSpeed(playbackSpeed); err != nil {
		return fmt.Errorf(errFormat, "failed to set playback speed", err)
	}

	if p.osdConfig.showOSD {
		if err := p.updateDisplay(p.speedState.current, playbackSpeed); err != nil {
			return fmt.Errorf(errFormat, errOSDUpdate, err)
		}
	}

	p.speedState.last = p.speedState.current

	return p.player.setPause(false)
}

// updateDisplay updates the on-screen display
func (p *PlaybackController) updateDisplay(cycleSpeed, playbackSpeed float64) error {

	if !p.osdConfig.showOSD {
		return nil
	}

	if cycleSpeed == 0 {
		return p.player.showOSDText("Paused")
	}

	var osdText strings.Builder

	if p.osdConfig.displayCycleSpeed {
		fmt.Fprintf(&osdText, "Cycle Speed: %.1f %s\n", cycleSpeed, p.speedConfig.SpeedUnits)
	}

	if p.osdConfig.displayPlaybackSpeed {
		fmt.Fprintf(&osdText, "Playback Speed: %.2fx\n", playbackSpeed)
	}

	if p.osdConfig.displayTimeRemaining {

		if timeRemaining, err := p.getTimeRemaining(); err == nil {
			fmt.Fprintf(&osdText, "Time Remaining: %s\n", formatSeconds(timeRemaining))
		} else {
			fmt.Fprintf(&osdText, "Time Remaining: %s\n", "????")
			logger.Warn(logger.VIDEO, errTimeRemaining+":", err)
		}

	}

	return p.player.showOSDText(osdText.String())
}

// getTimeRemaining calculates the time remaining in the video
func (p *PlaybackController) getTimeRemaining() (int64, error) {
	return p.player.getTimeRemaining()
}

// logDebugInfo logs debug information about current speeds
func (p *PlaybackController) logDebugInfo(speedController *speed.Controller) {

	logger.Debug(logger.VIDEO, "sensor speed buffer: ["+strings.Join(speedController.GetSpeedBuffer(), " ")+"]")
	logger.Debug(logger.VIDEO, logger.Magenta+"smoothed sensor speed:", strconv.FormatFloat(p.speedState.current, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"last playback speed:", strconv.FormatFloat(p.speedState.last, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"sensor speed delta:", strconv.FormatFloat(math.Abs(p.speedState.current-p.speedState.last), 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"playback speed update threshold:", strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64), p.speedConfig.SpeedUnits)

}

// formatSeconds converts seconds into HH:MM:SS format
func formatSeconds(seconds int64) string {

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
}
