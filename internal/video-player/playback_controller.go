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

// Common errors
var (
	errOSDUpdate     = fmt.Errorf("failed to update OSD")
	errPlaybackSpeed = fmt.Errorf("failed to set playback speed")
	errVideoComplete = fmt.Errorf("playback completed: normal exit")
)

const errFormat = "%w: %v"

// PlaybackController manages video playback using MPV media player
type PlaybackController struct {
	config      config.VideoConfig
	speedConfig config.SpeedConfig
	player      *mpv.Mpv
}

// speedState maintains the current state of playback speed
type speedState struct {
	current float64
	last    float64
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

// Start configures and starts the MPV media player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.SpeedController) error {

	logger.Debug(logger.VIDEO, "starting MPV video player...")
	defer p.player.TerminateDestroy()

	if err := p.setup(); err != nil {
		return fmt.Errorf("failed to setup player: %w", err)
	}

	return p.run(ctx, speedController)
}

// setup handles initial player configuration and video loading
func (p *PlaybackController) setup() error {

	if err := p.configureMPVPlayer(); err != nil {
		return fmt.Errorf("failed to configure MPV player: %w", err)
	}

	logger.Debug(logger.VIDEO, "loading video file:", p.config.FilePath)
	if err := p.player.Command([]string{"loadfile", p.config.FilePath}); err != nil {
		return fmt.Errorf("failed to load video file: %w", err)
	}

	return nil
}

// configureMPVPlayer sets up the player window based on configuration
func (p *PlaybackController) configureMPVPlayer() error {

	// keep open to allow for EOF check
	if err := p.player.SetOptionString("keep-open", "yes"); err != nil {
		return err
	}

	if err := p.player.SetOption("osd-font-size", mpv.FormatInt64, p.config.OnScreenDisplay.FontSize); err != nil {
		return err
	}

	if p.config.WindowScaleFactor == 1.0 {
		logger.Debug(logger.VIDEO, "maximizing video window")
		return p.player.SetOptionString("window-maximized", "yes")
	}

	scalePercent := strconv.Itoa(int(p.config.WindowScaleFactor * 100))
	return p.player.SetOptionString("autofit", scalePercent+"%")
}

// run handles the main playback loop
func (p *PlaybackController) run(ctx context.Context, speedController *speed.SpeedController) error {

	// Set an interval to check for updates
	ticker := time.NewTicker(time.Millisecond * time.Duration(p.config.UpdateIntervalSec*1000))
	defer ticker.Stop()

	state := &speedState{}
	logger.Info(logger.VIDEO, "MPV video playback started")

	for {
		select {
		case <-ctx.Done():
			fmt.Print("\r")
			logger.Info(logger.VIDEO, "user-generated interrupt, stopping MPV video player...")
			return nil
		case <-ticker.C:

			if err := p.tick(speedController, state); err != nil {

				if errors.Is(err, errVideoComplete) {
					return err
				}

				logger.Warn(logger.VIDEO, "playback error:", err.Error())
			}

		}
	}
}

// tick handles a single update cycle
func (p *PlaybackController) tick(speedController *speed.SpeedController, state *speedState) error {

	// First, check if playback is complete
	if complete, err := p.isPlaybackComplete(); err != nil || complete {
		return errVideoComplete
	}

	// Next, update the speed
	state.current = speedController.GetSmoothedSpeed()
	p.logDebugInfo(speedController, state)

	if state.current == 0 {
		return p.handleZeroSpeed()
	}

	if p.shouldUpdateSpeed(state) {
		return p.updateSpeed(state)
	}

	return nil
}

// isPlaybackComplete checks if the video has finished playing
func (p *PlaybackController) isPlaybackComplete() (bool, error) {

	reachedEOF, err := p.player.GetProperty("eof-reached", mpv.FormatFlag)
	if err != nil {
		return false, nil
	}

	return reachedEOF.(bool), nil
}

// shouldUpdateSpeed determines if the playback speed needs updating
func (p *PlaybackController) shouldUpdateSpeed(state *speedState) bool {
	return math.Abs(state.current-state.last) > p.speedConfig.SpeedThreshold
}

// handleZeroSpeed handles the case when no speed is detected
func (p *PlaybackController) handleZeroSpeed() error {

	logger.Debug(logger.VIDEO, "no speed detected, pausing video")

	if err := p.updateDisplay(0.0, 0.0); err != nil {
		return fmt.Errorf(errFormat, errOSDUpdate, err)
	}

	return p.player.SetProperty("pause", mpv.FormatFlag, true)
}

// updateSpeed adjusts the playback speed based on current speed
func (p *PlaybackController) updateSpeed(state *speedState) error {

	playbackSpeed := (state.current * p.config.SpeedMultiplier) / 10.0

	logger.Debug(logger.VIDEO, logger.Cyan+"updating video playback speed to",
		strconv.FormatFloat(playbackSpeed, 'f', 2, 64))

	if err := p.player.SetProperty("speed", mpv.FormatDouble, playbackSpeed); err != nil {
		return fmt.Errorf(errFormat, errPlaybackSpeed, err)
	}

	if err := p.updateDisplay(state.current, playbackSpeed); err != nil {
		return fmt.Errorf(errFormat, errOSDUpdate, err)
	}

	state.last = state.current
	return p.player.SetProperty("pause", mpv.FormatFlag, false)
}

// updateDisplay updates the on-screen display
func (p *PlaybackController) updateDisplay(cycleSpeed, playbackSpeed float64) error {

	if !p.config.OnScreenDisplay.ShowOSD {
		return nil
	}

	if cycleSpeed == 0 {
		return p.player.SetOptionString("osd-msg1", " Paused")
	}

	var osdText strings.Builder

	// Build the text string to display in OSD
	switch {
	case p.config.OnScreenDisplay.DisplayCycleSpeed:
		fmt.Fprintf(&osdText, " Cycle Speed: %.2f %s\n", cycleSpeed, p.speedConfig.SpeedUnits)
		fallthrough
	case p.config.OnScreenDisplay.DisplayPlaybackSpeed:
		fmt.Fprintf(&osdText, " Playback Speed: %.2fx\n", playbackSpeed)
		fallthrough
	case p.config.OnScreenDisplay.DisplayTimeRemaining:
		timeRemaining, _ := p.player.GetProperty("time-remaining", mpv.FormatInt64)
		fmt.Fprintf(&osdText, " Time Remaining: %s\n", formatSeconds(timeRemaining.(int64)))
	}

	return p.player.SetOptionString("osd-msg1", osdText.String())
}

// logDebugInfo logs debug information about current speeds
func (p *PlaybackController) logDebugInfo(speedController *speed.SpeedController, state *speedState) {
	logger.Debug(logger.VIDEO, "sensor speed buffer: ["+strings.Join(speedController.GetSpeedBuffer(), " ")+"]")
	logger.Debug(logger.VIDEO, logger.Magenta+"smoothed sensor speed:",
		strconv.FormatFloat(state.current, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"last playback speed:",
		strconv.FormatFloat(state.last, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"sensor speed delta:",
		strconv.FormatFloat(math.Abs(state.current-state.last), 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"playback speed update threshold:",
		strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64), p.speedConfig.SpeedUnits)
}

// FormatSeconds converts seconds into HH:MM:SS format
func formatSeconds(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
}
