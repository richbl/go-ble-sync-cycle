package video

import (
	"context"
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
	errOSDUpdate       = fmt.Errorf("failed to update OSD")
	errPlaybackSpeed   = fmt.Errorf("failed to set playback speed")
	errVideoComplete   = fmt.Errorf("MPV video playback completed")
	errGetVideoState   = fmt.Errorf("failed get media player state")
	errUnsupportedType = fmt.Errorf("unsupported type")
)

const (
	errTypeFormat = "%w: got %T"
	errFormat     = "%w: %v"
	osdMargin     = 40
)

// PlaybackController manages video playback using the MPV media player
type PlaybackController struct {
	videoConfig         config.VideoConfig
	speedConfig         config.SpeedConfig
	osdConfig           OSDConfig
	mpvPlayer           *mpv.Mpv
	speedState          *speedState
	speedUnitMultiplier float64
}

// speedState holds the state of the speedController speed
type speedState struct {
	current float64
	last    float64
}

// OSDConfig manages the configuration for the On-Screen Display (OSD)
type OSDConfig struct {
	ShowOSD              bool
	FontSize             int
	DisplayCycleSpeed    bool
	DisplayPlaybackSpeed bool
	DisplayTimeRemaining bool
}

// speedUnitConversion maps units of speed (mph, km/h) to their multiplier for consistent playback speed
var speedUnitConversion = map[string]float64{
	config.SpeedUnitsKMH: 1.60934,
	config.SpeedUnitsMPH: 1.0,
}

// NewPlaybackController creates a new mpv video player instance with the given config
func NewPlaybackController(videoConfig config.VideoConfig, speedConfig config.SpeedConfig) (*PlaybackController, error) {

	mpvPlayer := mpv.New()

	if err := mpvPlayer.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize MPV player: %w", err)
	}

	// Create and populate the OSDConfig struct
	osdConfig := OSDConfig{
		ShowOSD:              videoConfig.OnScreenDisplay.ShowOSD,
		FontSize:             videoConfig.OnScreenDisplay.FontSize,
		DisplayCycleSpeed:    videoConfig.OnScreenDisplay.DisplayCycleSpeed,
		DisplayPlaybackSpeed: videoConfig.OnScreenDisplay.DisplayPlaybackSpeed,
		DisplayTimeRemaining: videoConfig.OnScreenDisplay.DisplayTimeRemaining,
	}

	return &PlaybackController{
		videoConfig: videoConfig,
		speedConfig: speedConfig,
		osdConfig:   osdConfig,
		mpvPlayer:   mpvPlayer,
		speedState:  &speedState{},
	}, nil
}

// Start configures and starts playback of the MPV media player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.Controller) error {

	logger.Info(logger.VIDEO, "starting MPV video playback...")
	defer p.mpvPlayer.TerminateDestroy()

	// Configure the MPV media player
	if err := p.configurePlayback(); err != nil {
		return fmt.Errorf("failed to configure MPV video playback: %w", err)
	}

	// Start the event callback loop for the mpv media player
	if err := p.eventLoop(ctx, speedController); err != nil {
		return err
	}

	return nil
}

// configurePlayback sets up the player window based on configuration
func (p *PlaybackController) configurePlayback() error {

	if err := p.loadVideoFile(); err != nil {
		return err
	}

	if err := p.setMPVCallbackObservers(); err != nil {
		return err
	}

	if err := p.configurePlaybackWindow(); err != nil {
		return err
	}

	if err := p.configureKeepOpen(); err != nil {
		return err
	}

	if err := p.configureOSD(); err != nil {
		return err
	}

	if err := p.seekToStartPosition(); err != nil {
		return err
	}

	// Precalculate playback speed multiplier based on speed units
	p.speedUnitMultiplier = p.videoConfig.SpeedMultiplier / (speedUnitConversion[p.speedConfig.SpeedUnits] * 10.0)

	return nil
}

// loadVideoFile loads the video file into the mpv media player
func (p *PlaybackController) loadVideoFile() error {

	logger.Debug(logger.VIDEO, "loading video file:", p.videoConfig.FilePath)

	if err := p.mpvPlayer.Command([]string{"loadfile", p.videoConfig.FilePath}); err != nil {
		return fmt.Errorf("failed to load video file: %w", err)
	}

	return nil
}

// setMPVCallbackObservers sets mpv callback observers
func (p *PlaybackController) setMPVCallbackObservers() error {
	return p.mpvPlayer.ObserveProperty(0, "eof-reached", mpv.FormatFlag)
}

// configurePlaybackWindow sets up the playback window based on configuration
func (p *PlaybackController) configurePlaybackWindow() error {

	if p.videoConfig.WindowScaleFactor == 1.0 {
		if err := p.mpvPlayer.SetOptionString("fullscreen", "yes"); err != nil {
			return err
		}
	} else {
		scalePercent := strconv.Itoa(int(p.videoConfig.WindowScaleFactor * 100))
		if err := p.mpvPlayer.SetOptionString("autofit", scalePercent+"%"); err != nil {
			return err
		}
	}

	return nil
}

// configureKeepOpen ensures MPV keeps running after playback
func (p *PlaybackController) configureKeepOpen() error {

	if err := p.mpvPlayer.SetOptionString("keep-open", "yes"); err != nil {
		return err
	}

	return nil
}

// configureOSD sets up the OSD based on osdConfig struct
func (p *PlaybackController) configureOSD() error {

	if !p.osdConfig.ShowOSD {
		return nil
	}

	if err := p.mpvPlayer.SetOption("osd-font-size", mpv.FormatInt64, int64(p.osdConfig.FontSize)); err != nil {
		return err
	}

	if err := p.mpvPlayer.SetOption("osd-margin-x", mpv.FormatInt64, osdMargin); err != nil {
		return err
	}

	return nil
}

// seekToStartPosition seeks to the configured start position
func (p *PlaybackController) seekToStartPosition() error {

	logger.Debug(logger.VIDEO, "seeking to playback position", p.videoConfig.SeekToPosition)

	if err := p.mpvPlayer.SetOptionString("start", p.videoConfig.SeekToPosition); err != nil {
		return err
	}

	return nil
}

// eventLoop is the main event loop for the mpv media player
func (p *PlaybackController) eventLoop(ctx context.Context, speedController *speed.Controller) error {

	// Start a ticker to check updates from SpeedController
	ticker := time.NewTicker(time.Millisecond * time.Duration(p.videoConfig.UpdateIntervalSec*1000))
	defer ticker.Stop()

	for {
		select {

		// Check for updates from speedController
		case <-ticker.C:
			if err := p.updateSpeedFromController(speedController); err != nil {
				logger.Warn(logger.VIDEO, "speed update error:", err)
			}

		// Check for context cancellation
		case <-ctx.Done():
			fmt.Print("\r") // Clear the ^C character from the terminal line
			logger.Info(logger.VIDEO, "interrupt detected, stopping MPV video playback...")

			return nil
		}

		// Check for MPV events
		if err := p.handleMPVEvents(); err != nil {
			return err
		}
	}
}

// handleMPVEvents handles callback events from the mpv media player
func (p *PlaybackController) handleMPVEvents() error {

	event := p.mpvPlayer.WaitEvent(0)

	if event.EventID == mpv.EventPropertyChange {
		prop := event.Property()

		if prop.Name == "eof-reached" {
			reached, ok := prop.Data.(int)

			if ok && reached == 1 {
				return errVideoComplete
			}

		}
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

	return p.mpvPlayer.SetProperty("pause", mpv.FormatFlag, true)
}

// shouldUpdateSpeed determines if the playback speed needs updating
func (p *PlaybackController) shouldUpdateSpeed() bool {

	// Always update the speed if "display time remaining" option is enabled
	// Else update only if the speed delta is greater than the configured speed threshold
	return p.osdConfig.DisplayTimeRemaining ||
		(math.Abs(p.speedState.current-p.speedState.last) > p.speedConfig.SpeedThreshold)
}

// updateSpeed adjusts the playback speed based on current speed
func (p *PlaybackController) updateSpeed() error {

	// Update the playback speed based on current speed and unit multiplier
	playbackSpeed := p.speedState.current * p.speedUnitMultiplier

	logger.Debug(logger.VIDEO, logger.Cyan+"updating video playback speed to",
		strconv.FormatFloat(playbackSpeed, 'f', 2, 64)+"x")

	if err := p.mpvPlayer.SetProperty("speed", mpv.FormatDouble, playbackSpeed); err != nil {
		return fmt.Errorf(errFormat, errPlaybackSpeed, err)
	}

	if p.osdConfig.ShowOSD {
		if err := p.updateDisplay(p.speedState.current, playbackSpeed); err != nil {
			return fmt.Errorf(errFormat, errOSDUpdate, err)
		}
	}

	p.speedState.last = p.speedState.current

	return p.mpvPlayer.SetProperty("pause", mpv.FormatFlag, false)
}

// updateDisplay updates the on-screen display
func (p *PlaybackController) updateDisplay(cycleSpeed, playbackSpeed float64) error {

	if cycleSpeed == 0 {
		return p.mpvPlayer.SetOptionString("osd-msg1", "Paused")
	}

	var osdText strings.Builder

	if p.osdConfig.DisplayCycleSpeed {
		fmt.Fprintf(&osdText, "Cycle Speed: %.1f %s\n", cycleSpeed, p.speedConfig.SpeedUnits)
	}

	if p.osdConfig.DisplayPlaybackSpeed {
		fmt.Fprintf(&osdText, "Playback Speed: %.2fx\n", playbackSpeed)
	}

	if p.osdConfig.DisplayTimeRemaining {

		if timeRemaining, err := p.getTimeRemaining(); err == nil {
			fmt.Fprintf(&osdText, "Time Remaining: %s\n", formatSeconds(timeRemaining))
		} else {
			return fmt.Errorf(errFormat, errGetVideoState, err)
		}

	}

	return p.mpvPlayer.SetOptionString("osd-msg1", osdText.String())
}

// getTimeRemaining calculates the time remaining in the video
func (p *PlaybackController) getTimeRemaining() (int64, error) {

	timeRemaining, err := p.mpvPlayer.GetProperty("time-remaining", mpv.FormatInt64)
	if err != nil {
		return 0, fmt.Errorf(errFormat, errGetVideoState, err)
	}

	// Check if timeRemaining is an int64
	timeRemainingInt, ok := timeRemaining.(int64)
	if !ok {
		return 0, fmt.Errorf(errTypeFormat, errUnsupportedType, timeRemainingInt)
	}

	return timeRemainingInt, nil
}

// logDebugInfo logs debug information about current speeds
func (p *PlaybackController) logDebugInfo(speedController *speed.Controller) {

	logger.Debug(logger.VIDEO, "sensor speed buffer: ["+strings.Join(speedController.GetSpeedBuffer(), " ")+"]")
	logger.Debug(logger.VIDEO, logger.Magenta+"smoothed sensor speed:", strconv.FormatFloat(p.speedState.current, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"last playback speed:", strconv.FormatFloat(p.speedState.last, 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"sensor speed delta:", strconv.FormatFloat(math.Abs(p.speedState.current-p.speedState.last), 'f', 2, 64), p.speedConfig.SpeedUnits)
	logger.Debug(logger.VIDEO, logger.Magenta+"playback speed update threshold:", strconv.FormatFloat(p.speedConfig.SpeedThreshold, 'f', 2, 64), p.speedConfig.SpeedUnits)
}

// FormatSeconds converts seconds into HH:MM:SS format
func formatSeconds(seconds int64) string {

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
}
