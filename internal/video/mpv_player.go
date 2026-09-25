package video

/*
// DO NOT REMOVE: mpv player library expects C locale set to LC_NUMERIC:
//
#include <locale.h>
#include <stdlib.h>
static void set_c_locale_numeric() {
    setlocale(LC_NUMERIC, "C");
}
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	mpv "github.com/gen2brain/go-mpv"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// mpvPlayer is a wrapper around the go-mpv client
type mpvPlayer struct {
	player *mpv.Mpv
	mu     sync.RWMutex
}

// mpv-specific error definitions
var (
	errMPVPlayback = errors.New("mpv playback error")
)

// mpv timing constants
const (
	// maxFileLoadWait is the maximum time to wait for mpv to open and initialize
	// a loaded file (signalled by MPV_EVENT_FILE_LOADED) before declaring a timeout.
	// Applies to both the temporary headless validation instance and the playback
	// instance via the shared waitForFileLoaded method.
	maxFileLoadWait = 10 * time.Second

	// maxDimensionWait is the maximum time to wait for the video width/height
	// properties to become available after MPV_EVENT_FILE_LOADED. They are normally
	// available immediately; this covers rare demuxer metadata lag.
	maxDimensionWait = 2 * time.Second

	// dimensionPollInterval is the polling cadence used while waiting for video dimensions
	dimensionPollInterval = 50 * time.Millisecond
)

// newMpvPlayer creates a new mpvPlayer instance
func newMpvPlayer(ctx context.Context, videoConfig config.VideoConfig) (*mpvPlayer, error) {

	// Ensure C locale is set to "C" for numeric formats
	C.set_c_locale_numeric()

	player := mpv.New()
	if player == nil {
		return nil, errFailedToCreatePlayer
	}

	m := &mpvPlayer{player: player}

	// Attempt to force Wayland context if we detect a Wayland environment
	m.setupGPUContext(ctx)

	// Apply display targeting logic based on validation result
	if err := m.setupDisplayTargeting(ctx, videoConfig); err != nil {
		m.player.TerminateDestroy()
		m.player = nil

		return nil, err
	}

	// Initialize the mpv player
	if err := m.player.Initialize(); err != nil {
		m.player.TerminateDestroy()
		m.player = nil

		return nil, fmt.Errorf(errFormat, "failed to initialize mpv player", err)
	}

	logger.Info(ctx, logger.VIDEO, "mpv player object created")

	return m, nil
}

// setupGPUContext attempts to force Wayland context if we detect a Wayland environment
func (m *mpvPlayer) setupGPUContext(ctx context.Context) {

	if os.Getenv("WAYLAND_DISPLAY") != "" {

		if err := m.player.SetOptionString("gpu-context", "wayland"); err != nil {
			logger.Warn(ctx, logger.VIDEO, fmt.Sprintf("failed to set gpu-context=wayland: %v", err))
		} else {
			logger.Debug(ctx, logger.VIDEO, "mpv configured for native Wayland context")
		}

	}

}

// setupDisplayTargeting configures mpv to target a specific display
func (m *mpvPlayer) setupDisplayTargeting(ctx context.Context, videoConfig config.VideoConfig) error {

	targetName := videoConfig.ValidationResult.ActualDisplayName
	isValid := videoConfig.ValidationResult.IsValid
	isNonDefault := videoConfig.ValidationResult.IsNonDefaultMonitor

	// If we have a valid non-default monitor, target it directly in fullscreen mode
	if isValid && isNonDefault {

		// Force fullscreen, which is required to lock to a Wayland output
		if err := m.player.SetOptionString("fs", "yes"); err != nil {
			return fmt.Errorf("failed to set fullscreen (fs) option: %w", err)
		}

		// Target the specific hardware connector
		if err := m.player.SetOptionString("fs-screen-name", targetName); err != nil {
			return fmt.Errorf("failed to set fs-screen-name option to %s: %w", targetName, err)
		}

		logger.Info(ctx, logger.VIDEO, "mpv configured to target non-default display in fullscreen: "+targetName)

		return nil
	}

	// Either default/embedded monitor or invalid/fallback: Wayland supports windowed mode here
	if err := m.player.SetOptionString("fs", "no"); err != nil {
		return fmt.Errorf("failed to unset fullscreen option: %w", err)
	}

	if isValid {
		logger.Info(ctx, logger.VIDEO, "mpv configured to target default display in windowed mode: "+targetName)
	} else {

		if videoConfig.TargetDisplayName != "" {
			logger.Warn(ctx, logger.VIDEO, fmt.Sprintf("target display '%s' not found; falling back to default display in windowed mode", videoConfig.TargetDisplayName))
		} else {
			logger.Info(ctx, logger.VIDEO, "no target display specified; using default display in windowed mode")
		}
	}

	return nil
}

// validateVideoFile validates the video file using a temporary headless MPV instance.
//
// Validation is event-driven rather than property-polled: mpv signals a successful
// load with MPV_EVENT_FILE_LOADED, and definitive load failures (missing file,
// unsupported or corrupt container, decode failure) with MPV_EVENT_END_FILE, which
// carries the actual error in its payload. There is no pollable "error" property
// in mpv — the error data lives exclusively in the end-file event. Do not
// reintroduce property polling here.
func (m *mpvPlayer) validateVideoFile(videoPath, position string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {

		tempMpv := mpv.New()
		if tempMpv == nil {
			return errFailedToCreatePlayer
		}

		defer tempMpv.TerminateDestroy()

		// Configure and initialize for headless operation
		if err := configureHeadless(tempMpv); err != nil {
			return err
		}

		// Load the file paused: mpv only needs to demux and probe streams,
		// not decode or advance frames
		if err := tempMpv.Command([]string{"loadfile", videoPath, "replace", "0", "pause=yes"}); err != nil {
			return fmt.Errorf(errFormat, errFailedToLoadVideo, err)
		}

		// Wait for a definitive outcome: loaded, or failed with mpv's actual error
		if err := m.waitForFileLoaded(tempMpv); err != nil {
			return err
		}

		// MPV_EVENT_FILE_LOADED has fired: track selection is complete and
		// per-file properties (video-codec, width, height, duration) are
		// reliably available
		if err := m.validateVideoTrack(tempMpv); err != nil {
			return err
		}

		return m.validateSeekPosition(tempMpv, position)

	})
}

// configureHeadless configures an mpv instance for headless validation operation.
// This is a package-level function because it operates solely on its parameter
// and does not require access to the mpvPlayer receiver.
func configureHeadless(p *mpv.Mpv) error {

	opts := []struct{ key, value string }{
		{"vo", "null"},
		{"ao", "null"},
		{"ytdl", "no"},
	}

	for _, opt := range opts {

		if err := p.SetOptionString(opt.key, opt.value); err != nil {
			return fmt.Errorf("failed to set option %s: %w", opt.key, err)
		}

	}

	if err := p.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize headless mpv: %w", err)
	}

	return nil
}

// waitForFileLoaded waits for a definitive file-load outcome from an mpv instance:
// MPV_EVENT_FILE_LOADED on success, or MPV_EVENT_END_FILE (with the actual error)
// on failure. This single implementation serves both the temporary headless
// validation instance and the playback instance.
//
// Note: loadfile is asynchronous — its return code only indicates that the command
// was queued, so the outcome must be learned from the event queue.
func (m *mpvPlayer) waitForFileLoaded(p *mpv.Mpv) error {

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "waiting for mpv file-loaded event...")

	timeout := time.After(maxFileLoadWait)

	for {

		select {
		case <-timeout:
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("timeout after %s waiting for file to load", maxFileLoadWait))

			return errMediaParseTimeout

		default:

			event := p.WaitEvent(0.1)
			if event == nil || event.EventID == mpv.EventNone {
				continue
			}

			switch event.EventID {

			case mpv.EventFileLoaded:
				logger.Debug(logger.BackgroundCtx, logger.VIDEO, "mpv file-loaded event received")

				// Discard any residual load-phase events so subsequent event
				// processing starts from a clean queue (no-op for the
				// temporary headless instance, which is destroyed after validation)
				m.drainEvents(p)

				return nil

			case mpv.EventEnd:
				// mpv reports load failures here, with the concrete error in
				// the event payload
				return m.handleEndFile(event)

			case mpv.EventShutdown:
				return fmt.Errorf("%w: mpv core shut down while loading file", errFailedToLoadVideo)
			}

		}

	}

}

// handleEndFile processes an MPV_EVENT_END_FILE event received while waiting for
// a file to load. Since EventFileLoaded has not yet been observed, every end-file
// reason is treated as a load failure. Each reason produces a descriptive error
// wrapping errFailedToLoadVideo so callers can match with errors.Is.
//
// Note: a file that ends during normal playback is not routed here; that case is
// detected by the controller's event loop (eof-reached property / end-file event)
// and reported as ErrVideoComplete.
func (m *mpvPlayer) handleEndFile(event *mpv.Event) error {

	endFile := event.EndFile()

	logger.Debug(logger.BackgroundCtx, logger.VIDEO,
		fmt.Sprintf("mpv end-file event during load: reason=%v error=%v", endFile.Reason, endFile.Error))

	// A concrete error from mpv takes precedence over the reason code
	if endFile.Error != nil {
		return fmt.Errorf("%w: %w", errFailedToLoadVideo, endFile.Error)
	}

	switch endFile.Reason {
	case mpv.EndFileEOF:
		return fmt.Errorf("%w: file ended before mpv reported it loaded", errFailedToLoadVideo)
	case mpv.EndFileStop:
		return fmt.Errorf("%w: file load was stopped", errFailedToLoadVideo)
	case mpv.EndFileQuit:
		return fmt.Errorf("%w: mpv quit while loading file", errFailedToLoadVideo)
	case mpv.EndFileRedirect:
		return fmt.Errorf("%w: file load was redirected before completion", errFailedToLoadVideo)
	case mpv.EndFileError:
		return fmt.Errorf("%w: mpv reported a load error", errFailedToLoadVideo)
	default:
		return fmt.Errorf("%w: unexpected end-file reason %d", errFailedToLoadVideo, endFile.Reason)
	}

}

// validateVideoTrack confirms a loaded file contains a playable video track with
// valid dimensions. It must be called after MPV_EVENT_FILE_LOADED, when track
// selection is complete and per-file properties are available.
func (m *mpvPlayer) validateVideoTrack(p *mpv.Mpv) error {

	// After FILE_LOADED, an unavailable or empty video-codec reliably means
	// "no playable video track" (e.g. an audio-only file)
	vCodec, err := p.GetProperty("video-codec", mpv.FormatString)
	if err != nil || !isNonEmptyString(vCodec) {
		return errNoVideoTrack
	}

	// Dimensions normally accompany the codec but can lag in rare cases:
	// wait briefly for them to settle
	info := m.waitForDimensions(p)
	if info == nil || info.width <= 0 || info.height <= 0 {
		return errInvalidVideoDimensions
	}

	logger.Debug(logger.BackgroundCtx, logger.VIDEO,
		fmt.Sprintf("video track validated: codec=%s dimensions=%dx%d", vCodec, info.width, info.height))

	return nil
}

// waitForDimensions polls briefly for the video width/height properties to become
// available, returning nil if they never do
func (m *mpvPlayer) waitForDimensions(p *mpv.Mpv) *videoValidationInfo {

	timeout := time.After(maxDimensionWait)
	ticker := time.NewTicker(dimensionPollInterval)
	defer ticker.Stop()

	for {

		select {
		case <-timeout:
			return nil

		case <-ticker.C:
			if info, ok := m.extractStreamInfo(p); ok && info.width > 0 && info.height > 0 {
				return info
			}
		}

	}

}

// extractStreamInfo extracts the video codec and dimensions from an mpv instance.
// It reports false when the video codec is not (yet) available; dimension values
// may still be zero while demuxer metadata settles.
func (m *mpvPlayer) extractStreamInfo(p *mpv.Mpv) (*videoValidationInfo, bool) {

	vCodec, err := p.GetProperty("video-codec", mpv.FormatString)
	if err != nil || !isNonEmptyString(vCodec) {
		return nil, false
	}

	info := &videoValidationInfo{}

	if val, err := p.GetProperty("width", mpv.FormatInt64); err == nil {
		if width, ok := val.(int64); ok {
			info.width = int(width)
		}
	}

	if val, err := p.GetProperty("height", mpv.FormatInt64); err == nil {
		if height, ok := val.(int64); ok {
			info.height = int(height)
		}
	}

	return info, true
}

// validateSeekPosition checks if the requested seek position is within the video
// duration. A zero seek position is always valid and skips the duration query.
//
// mpv's "duration" property is natively a double (seconds as floating point);
// requesting FormatDouble avoids the precision loss that FormatInt64 can introduce
// for fractional-second durations.
func (m *mpvPlayer) validateSeekPosition(p *mpv.Mpv, position string) error {

	// Parse requested seek position in milliseconds
	seekPosition, err := parseTimePosition(position)
	if err != nil {
		return fmt.Errorf(errFormat, "unable to parse specified seek time", err)
	}

	// 0 seek position is always valid as long as the file loaded successfully
	if seekPosition == 0 {
		return nil
	}

	// Get playback duration in milliseconds (mpv duration is natively double/seconds)
	val, err := p.GetProperty("duration", mpv.FormatDouble)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	var duration int64

	switch v := val.(type) {
	case float64:
		duration = int64(v * 1000)
	case int64:
		duration = v * 1000
	default:
		return fmt.Errorf("%w: unexpected video duration property type: %T", errInvalidTimeFormat, val)
	}

	if int64(seekPosition) > duration {
		return fmt.Errorf("start/seek time (%ds) exceeds the video playback duration (%ds): %w",
			seekPosition/1000, duration/1000, ErrSeekExceedsDuration)
	}

	return nil
}

// loadFile loads a video file into the mpv player
func (m *mpvPlayer) loadFile(path string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "attempting to load file: "+path)

		if err := m.player.Command([]string{"loadfile", path, "replace", "0", "pause=yes"}); err != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("mpv command failed: %v", err))

			return wrapError(errFailedToLoadVideo.Error(), err)
		}

		// Wait for the file to load, or fail, via the shared event-driven wait
		if err := m.waitForFileLoaded(m.player); err != nil {
			return err
		}

		return nil
	})
}

// setSpeed sets the playback speed of the video
func (m *mpvPlayer) setSpeed(speed float64) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		return wrapError("failed to set video playback speed", m.player.SetProperty("speed", mpv.FormatDouble, speed))
	})
}

// setPause sets the pause state of the video
func (m *mpvPlayer) setPause(paused bool) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		return wrapError("failed to pause video", m.player.SetProperty("pause", mpv.FormatFlag, paused))
	})
}

// getInt64Property is a helper to retrieve an integer property from the mpv player
func (m *mpvPlayer) getInt64Property(property string, format mpv.Format, errorContext string) (int64, error) {

	return queryGuarded(&m.mu, func() bool { return m.player == nil }, func() (int64, error) {

		val, err := m.player.GetProperty(property, format)
		if err != nil {
			return 0, fmt.Errorf(errFormat, errorContext, err)
		}

		switch v := val.(type) {
		case int64:
			return v, nil
		case float64:
			return int64(v), nil
		default:
			return 0, errInvalidTimeFormat
		}
	})
}

// timeRemaining gets the remaining time of the video
func (m *mpvPlayer) timeRemaining() (int64, error) {
	return m.getInt64Property("time-remaining", mpv.FormatInt64, "failed to get video time remaining")
}

// playbackPosition gets the current elapsed time of the video
func (m *mpvPlayer) playbackPosition() (int64, error) {
	return m.getInt64Property("time-pos", mpv.FormatDouble, "failed to get video playback position")
}

// setPlaybackSize sets media player window size
func (m *mpvPlayer) setPlaybackSize(windowSize float64) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {

		// Enable fullscreen if window size is 1.0 (100%)
		if windowSize == 1.0 {
			return wrapError("failed to enable fullscreen", m.player.SetOptionString("fullscreen", "yes"))
		}

		// Scale video window size
		scaleValue := int(windowSize * 100)

		return wrapError("failed to set window size", m.player.SetOptionString("autofit", fmt.Sprintf("%d%%x%d%%", scaleValue, scaleValue)))
	})
}

// setKeepOpen configures the player to keep the window open after playback completes
func (m *mpvPlayer) setKeepOpen(keepOpen bool) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		value := "no"
		if keepOpen {
			value = "yes"
		}

		return wrapError("failed to set keep-open media player option", m.player.SetOptionString("keep-open", value))
	})
}

// seek moves the playback position to the specified time position
func (m *mpvPlayer) seek(position string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		return wrapError(errUnableToSeek.Error(), m.player.SetPropertyString("start", position))
	})
}

// setOSD configures the On-Screen Display (OSD)
func (m *mpvPlayer) setOSD(options osdConfig) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {

		if err := m.player.SetOption("osd-font-size", mpv.FormatInt64, int64(options.fontSize)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD font size", err)
		}

		if err := m.player.SetOption("osd-margin-x", mpv.FormatInt64, int64(options.marginX)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD horizontal margin", err)
		}

		if err := m.player.SetOption("osd-margin-y", mpv.FormatInt64, int64(options.marginY)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD vertical margin", err)
		}

		if err := m.player.SetOptionString("osd-align-x", options.alignX); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD horizontal position", err)
		}

		if err := m.player.SetOptionString("osd-align-y", options.alignY); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD vertical position", err)
		}

		return nil
	})
}

// setupEvents prepares the player to listen for end-of-file and file-loaded events
func (m *mpvPlayer) setupEvents() error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		return wrapError("failed to setup end-of-file observe event", m.player.ObserveProperty(0, "eof-reached", mpv.FormatFlag))
	})
}

// waitEvent waits for an mpv event and translates it to a generic playerEvent
func (m *mpvPlayer) waitEvent(timeout float64) *playerEvent {

	res, err := queryGuarded(&m.mu, func() bool { return m.player == nil }, func() (*playerEvent, error) {

		// If no event generated before timeout, return an empty event
		e := m.player.WaitEvent(timeout)
		if e == nil || e.EventID == mpv.EventNone {
			return &playerEvent{id: eventNone}, nil
		}

		switch e.EventID {

		case mpv.EventPropertyChange:
			// "eof-reached" is the reliable completion signal while keep-open=yes
			// is active, since MPV_EVENT_END_FILE never fires in that mode
			prop := e.Property()
			if prop.Name == "eof-reached" && isTrueFlag(prop.Data) {
				return &playerEvent{id: eventEndFile}, nil
			}

		case mpv.EventEnd:
			return &playerEvent{id: eventEndFile}, nil
		}

		return &playerEvent{id: eventNone}, nil
	})

	if err != nil {
		return nil
	}

	if res == nil {
		return &playerEvent{id: eventNone}
	}

	return res
}

// isTrueFlag reports whether an observed property value represents boolean true,
// tolerating the different Go types go-mpv may use to surface MPV_FORMAT_FLAG
// data across versions (bool, int, int64)
func isTrueFlag(data any) bool {

	switch v := data.(type) {
	case bool:
		return v
	case int:
		return v == 1
	case int64:
		return v == 1
	default:
		return false
	}

}

// showOSDText displays text on the OSD
func (m *mpvPlayer) showOSDText(text string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {
		return wrapError("failed to show OSD text", m.player.SetOptionString("osd-msg1", text))
	})
}

// terminatePlayer terminates the mpv player instance and cleans up resources
func (m *mpvPlayer) terminatePlayer() {

	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "starting player termination")

	if m.player != nil {

		// Run TerminateDestroy in a goroutine with timeout to prevent blocking
		done := make(chan struct{})
		go func() {
			m.player.TerminateDestroy()
			close(done)
		}()

		// Wait with timeout
		select {
		case <-done:
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "call to terminate mpv completed successfully")

		case <-time.After(2 * time.Second):
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, "call to terminate mpv timed out after 2s, continuing mpv shutdown")
		}

		m.player = nil
		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "destroyed MPV handle")
	}
}

// drainEvents discards any pending events from an mpv instance's event queue.
// The target handle is passed explicitly so the drain never depends on (or races
// with) the receiver's player state. WaitEvent(0) is non-blocking: only events
// already queued are discarded.
func (m *mpvPlayer) drainEvents(p *mpv.Mpv) {

	for range 10 {
		event := p.WaitEvent(0)
		if event == nil || event.EventID == mpv.EventNone {
			return
		}
	}

}

// isNonEmptyString checks if a property value is a non-empty string
func isNonEmptyString(prop any) bool {

	if prop == nil {
		return false
	}

	val, ok := prop.(string)

	return ok && val != ""
}
