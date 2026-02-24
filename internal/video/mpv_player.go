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
	"sync"
	"time"

	mpv "github.com/gen2brain/go-mpv"
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

// newMpvPlayer creates a new mpvPlayer instance
func newMpvPlayer(ctx context.Context) (*mpvPlayer, error) {

	// Ensure C locale is set to "C" for numeric formats
	C.set_c_locale_numeric()

	m := &mpvPlayer{
		player: mpv.New(),
	}

	if err := m.player.Initialize(); err != nil {
		return nil, fmt.Errorf(errFormat, "failed to initialize mpv player", err)
	}

	logger.Info(ctx, logger.VIDEO, "mpv player object created")

	return m, nil
}

// loadFile loads a video file into the mpv player
func (m *mpvPlayer) loadFile(path string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "attempting to load file: "+path)

		// Validate video file format with tmp (headless) mpv instance
		if err := m.validateVideoFile(path); err != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("video file validation failed: %v", err))

			return err
		}

		logger.Info(logger.BackgroundCtx, logger.VIDEO, "video file validation succeeded")

		// With file validated, load into full mpv player
		if err := m.player.Command([]string{"loadfile", path}); err != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("mpv command failed: %v", err))

			return wrapError(errFailedToLoadVideo.Error(), err)
		}

		// Wait for file to load in main player
		if err := m.waitForFileLoaded(); err != nil {
			return err
		}

		return nil
	})
}

// validateVideoFile validates the video file using a tmp/headless MPV instance
func (m *mpvPlayer) validateVideoFile(videoPath string) error {

	tempMpv := mpv.New()
	if tempMpv == nil {
		return errFailedToCreatePlayer
	}

	defer tempMpv.TerminateDestroy()

	// Configure for tmp/headless operation
	if err := m.configureHeadless(tempMpv); err != nil {
		return err
	}

	// Load file
	if err := tempMpv.Command([]string{"loadfile", videoPath}); err != nil {
		return fmt.Errorf(errFormat, errFailedToLoadVideo, err)
	}

	// Poll for active stream and then extract validation info
	info, err := m.pollForActiveStream(tempMpv)
	if err != nil {
		return err
	}

	// Get playback duration
	val, _ := tempMpv.GetProperty("duration", mpv.FormatDouble)
	if dur, ok := val.(float64); ok {
		info.duration = float64(dur)
	}

	// Validate the extracted information
	if info.width == 0 || info.height == 0 {
		return errInvalidVideoDimensions
	}

	return nil
}

// configureHeadless configures an mpv instance for tmp/headless operation
func (m *mpvPlayer) configureHeadless(p *mpv.Mpv) error {

	opts := map[string]string{
		"vo":   "null",
		"ao":   "null",
		"ytdl": "no",
	}

	// Set all mpv options
	for k, v := range opts {

		if err := p.SetOptionString(k, v); err != nil {
			return fmt.Errorf("failed to set option %s: %w", k, err)
		}

	}

	if err := p.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize headless MPV: %w", err)
	}

	return nil
}

// pollForActiveStream waits for video codec to appear and extracts stream information
func (m *mpvPlayer) pollForActiveStream(p *mpv.Mpv) (*videoValidationInfo, error) {

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Poll for video codec and dimensions to become available, with timeout
	for {

		select {
		case <-timeout:
			return nil, errStreamTimeout

		case <-ticker.C:

			// Check for internal errors
			if err := m.checkPlayerError(p); err != nil {
				return nil, err
			}

			// Attempt to extract stream info
			info, active := m.extractStreamInfo(p)
			if !active {
				continue
			}

			// If we have video codec but dimensions aren't ready yet, keep waiting
			if info.width == 0 && info.height == 0 {

				vCodec, _ := p.GetProperty("video-codec", mpv.FormatString)
				if isNonEmptyString(vCodec) {

					// Video track exists but dimensions not yet loaded
					continue
				}

				// No video codec
				return nil, errNoVideoTrack
			}

			return info, nil
		}
	}

}

// checkPlayerError checks if the MPV player has encountered an error
func (m *mpvPlayer) checkPlayerError(p *mpv.Mpv) error {

	errProp, _ := p.GetProperty("error", mpv.FormatString)
	if errProp != nil {

		if errMsg, ok := errProp.(string); ok && errMsg != "" {
			return fmt.Errorf(errFormat, errMsg, errMPVPlayback)
		}

	}

	return nil
}

// extractStreamInfo extracts video codec information and dimensions from the player
func (m *mpvPlayer) extractStreamInfo(p *mpv.Mpv) (*videoValidationInfo, bool) {

	vCodec, _ := p.GetProperty("video-codec", mpv.FormatString)
	hasCodec := isNonEmptyString(vCodec)

	if !hasCodec {
		return nil, false
	}

	info := &videoValidationInfo{}

	if hasCodec {

		val, _ := p.GetProperty("width", mpv.FormatInt64)
		if width, ok := val.(int64); ok {
			info.width = int(width)
		}

		val, _ = p.GetProperty("height", mpv.FormatInt64)
		if height, ok := val.(int64); ok {
			info.height = int(height)
		}

	}

	return info, true
}

// waitForFileLoaded waits for the file-loaded event in the mpv player
func (m *mpvPlayer) waitForFileLoaded() error {

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "waiting for file to load into mpv player")

	const maxEvents = 20
	eventCount := 0

	for eventCount < maxEvents {

		event := m.player.WaitEvent(0.5)
		if event == nil {
			eventCount++

			continue
		}

		switch event.EventID {
		case mpv.EventFileLoaded:
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "media file successfully loaded into mpv player")
			m.drainEvents()

			return nil

		case mpv.EventEnd:
			return m.handleEndFile(event)

		default:
			eventCount++

			continue
		}
	}

	logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("timeout after processing %d events", maxEvents))

	return errMediaParseTimeout
}

// handleEndFile processes the EventEndFile event during file loading
func (m *mpvPlayer) handleEndFile(event *mpv.Event) error {

	endFile := event.EndFile()

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("mpv EndFile event trigger: %v", endFile.Error))

	var validationErr error

	if endFile.Error != nil {
		validationErr = fmt.Errorf(errFormat, errFailedToLoadVideo, endFile.Error)
	} else {

		switch endFile.Reason {
		case mpv.EndFileError:
			validationErr = ErrVideoComplete

		case mpv.EndFileEOF:
			validationErr = errInvalidVideoDimensions

		default:
			validationErr = errPlaybackEndedUnexpectedly
		}

	}

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "draining remaining events after error")
	m.drainEvents()

	return validationErr
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

// timeRemaining gets the remaining time of the video
func (m *mpvPlayer) timeRemaining() (int64, error) {

	return queryGuarded(&m.mu, func() bool { return m.player == nil }, func() (int64, error) {
		if m.player == nil {
			return 0, errPlayerNotInitialized
		}

		timeRemaining, err := m.player.GetProperty("time-remaining", mpv.FormatInt64)
		if err != nil {
			return 0, fmt.Errorf(errFormat, "failed to get video time remaining", err)
		}

		timeRemainingInt, ok := timeRemaining.(int64)
		if !ok {
			return 0, errInvalidTimeFormat
		}

		return timeRemainingInt, nil
	})
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

		if err := m.player.SetOption("osd-margin-x", mpv.FormatInt64, int64(options.marginX)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD X position", err)
		}
		if err := m.player.SetOption("osd-margin-y", mpv.FormatInt64, int64(options.marginY)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD Y position", err)
		}
		if err := m.player.SetOption("osd-font-size", mpv.FormatInt64, int64(options.fontSize)); err != nil {
			return fmt.Errorf(errFormat, "failed to set OSD font size", err)
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

	res, _ := queryGuarded(&m.mu, func() bool { return m.player == nil }, func() (*playerEvent, error) {

		// If no event generated before timeout, return an empty event
		e := m.player.WaitEvent(timeout)
		if e == nil || e.EventID == mpv.EventNone {
			return &playerEvent{id: eventNone}, nil
		}

		switch e.EventID {
		case mpv.EventPropertyChange:
			prop := e.Property()
			if prop.Name == "eof-reached" {

				if val, ok := prop.Data.(int); ok && val == 1 {
					return &playerEvent{id: eventEndFile}, nil
				}

			}

		case mpv.EventEnd:
			return &playerEvent{id: eventEndFile}, nil
		}

		return &playerEvent{id: eventNone}, nil
	})

	if res == nil {
		return &playerEvent{id: eventNone}
	}

	return res
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
		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "destroyed MPV handle: C resources released")
	}
}

// drainEvents drains any remaining events from MPV's event queue
func (m *mpvPlayer) drainEvents() {

	for range 10 {
		event := m.player.WaitEvent(0.1)
		if event == nil {
			break
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
