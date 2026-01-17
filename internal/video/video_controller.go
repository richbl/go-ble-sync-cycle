package video

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/speed"
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
		return nil, errUnsupportedVideoPlayer
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

// StartPlayback configures and starts playback of the media player
func (p *PlaybackController) StartPlayback(ctx context.Context, speedController *speed.Controller) error {

	logger.Info(ctx, logger.VIDEO, fmt.Sprintf("starting %s video playback...", p.videoConfig.MediaPlayer))

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

// TimeRemaining is a thread-safe wrapper to get the video time remaining
func (p *PlaybackController) TimeRemaining() (string, error) {

	seconds, err := p.player.timeRemaining()
	if err != nil {
		return "--:--:--", err
	}

	return formatSeconds(seconds), nil
}

// PlaybackSpeed returns the current calculated playback rate multiplier
func (p *PlaybackController) PlaybackSpeed() float64 {

	if p.speedState == nil {
		return 0.0
	}

	return p.speedState.current * p.speedUnitMultiplier
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

			if err := p.updateSpeedFromController(ctx, speedController); err != nil {
				logger.Warn(ctx, logger.VIDEO, fmt.Sprintf("speed update error: %v", err))
			}

		case <-ctx.Done():

			fmt.Fprint(os.Stdout, "\r") // Clear the ^C character from the terminal line
			logger.Info(ctx, logger.VIDEO, fmt.Sprintf("interrupt detected, stopping %s video playback...", p.videoConfig.MediaPlayer))

			return nil // No need to show this context cancellation error
		}
	}

}

// handlePlayerEvents handles callback events from the media player
func (p *PlaybackController) handlePlayerEvents() error {

	event := p.player.waitEvent(0) // Use a non-blocking wait
	if event != nil && event.id == eventEndFile {
		return fmt.Errorf("%w", ErrVideoComplete)
	}

	return nil
}

// updateSpeedFromController manages updates from the speedController component
func (p *PlaybackController) updateSpeedFromController(ctx context.Context, speedController *speed.Controller) error {

	p.speedState.current = speedController.SmoothedSpeed()
	p.logDebugInfo(ctx, speedController)

	if p.speedState.current == 0 {
		return p.handleZeroSpeed(ctx)
	}

	if p.shouldUpdateSpeed() {
		return p.updateSpeed(ctx)
	}

	return nil
}

// handleZeroSpeed handles the case when no speed is detected
func (p *PlaybackController) handleZeroSpeed(ctx context.Context) error {

	logger.Debug(ctx, logger.VIDEO, "no speed detected, pausing video")

	if err := p.updateDisplay(ctx, 0.0, 0.0); err != nil {
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
func (p *PlaybackController) updateSpeed(ctx context.Context) error {

	// Update the playback speed based on current speed and unit multiplier
	playbackSpeed := p.PlaybackSpeed()

	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf(logger.Cyan+"updating video playback speed to %.2fx...", playbackSpeed))

	if err := p.player.setSpeed(playbackSpeed); err != nil {
		return fmt.Errorf(errFormat, "failed to set playback speed", err)
	}

	if p.osdConfig.showOSD {
		if err := p.updateDisplay(ctx, p.speedState.current, playbackSpeed); err != nil {
			return fmt.Errorf(errFormat, errOSDUpdate, err)
		}
	}

	p.speedState.last = p.speedState.current

	return p.player.setPause(false)
}

// updateDisplay updates the on-screen display
func (p *PlaybackController) updateDisplay(ctx context.Context, cycleSpeed, playbackSpeed float64) error {

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

		if timeRemaining, err := p.timeRemaining(); err == nil {
			fmt.Fprintf(&osdText, "Time Remaining: %s\n", formatSeconds(timeRemaining))
		} else {
			fmt.Fprintf(&osdText, "Time Remaining: %s\n", "????")
			logger.Warn(ctx, logger.VIDEO, fmt.Sprintf("%s: %v", errTimeRemaining, err))
		}

	}

	return p.player.showOSDText(osdText.String())
}

// timeRemaining calculates the time remaining in the video
func (p *PlaybackController) timeRemaining() (int64, error) {
	return p.player.timeRemaining()
}

// logDebugInfo logs debug information about current speeds
func (p *PlaybackController) logDebugInfo(ctx context.Context, speedController *speed.Controller) {

	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf("sensor speed buffer: [%s]", strings.Join(speedController.SpeedBuffer(ctx), " ")))
	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf(logger.Magenta+"smoothed sensor speed: %.2f %s", p.speedState.current, p.speedConfig.SpeedUnits))
	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf(logger.Magenta+"last playback speed: %.2f %s", p.speedState.last, p.speedConfig.SpeedUnits))
	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf(logger.Magenta+"sensor speed delta: %.2f %s", math.Abs(p.speedState.current-p.speedState.last), p.speedConfig.SpeedUnits))
	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf(logger.Magenta+"playback speed update threshold: %.2f %s", p.speedConfig.SpeedThreshold, p.speedConfig.SpeedUnits))
}

// formatSeconds converts seconds into HH:MM:SS format
func formatSeconds(seconds int64) string {

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
}
