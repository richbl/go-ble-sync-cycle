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

// newMpvPlayer creates a new mpvPlayer instance
func newMpvPlayer() (*mpvPlayer, error) {

	// Ensure C locale is set to "C" for numeric formats
	C.set_c_locale_numeric()

	m := &mpvPlayer{
		player: mpv.New(),
	}

	if err := m.player.Initialize(); err != nil {
		return nil, fmt.Errorf(errFormat, "failed to initialize mpv player", err)
	}

	return m, nil
}

// loadFile loads a video file into the mpv player and validates it
func (m *mpvPlayer) loadFile(path string) error {

	return execGuarded(&m.mu, func() bool { return m.player == nil }, func() error {

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "attempting to load file: "+path)

		if err := m.player.Command([]string{"loadfile", path}); err != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("mpv command failed: %v", err))

			return wrapError(errFailedToLoadVideo.Error(), err)
		}

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "command succeeded, now validating file...")

		// Wait for file-loaded event and then validate the file
		err := m.validateLoadedFile()
		if err != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("video file validation failed: %v", err))

			return err
		}

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "video file validation succeeded")

		return nil
	})
}

// validateLoadedFile waits for the file-loaded event and checks if video dimensions are valid
func (m *mpvPlayer) validateLoadedFile() error {

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "starting mpv video file validation loop")

	// Set a maximum number of events to process before giving up
	const maxEvents = 20
	eventCount := 0
	var validationErr error

	// Wait for the file-loaded event
	for eventCount < maxEvents {
		event := m.player.WaitEvent(0.5)
		if event == nil {
			eventCount++

			continue
		}

		switch event.EventID {
		case mpv.EventFileLoaded:

			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "media file successfully loaded: validating dimensions")

			// File loaded successfully, now validate dimensions
			if err := m.checkVideoDimensions(); err != nil {
				validationErr = err

				break
			}
			// Validation successful: drain any remaining events before returning
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "validation successful, draining remaining events")

			m.drainEvents()

			return nil

			// During file loading, a premature end event means the file is invalid
		case mpv.EventEnd:
			return m.handleEndFile(event)

		default:
			// Ignore all other events and continue waiting
			eventCount++

			continue
		}

	}

	if validationErr != nil {
		return validationErr
	}

	logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("timeout after processing %d events", maxEvents))

	return errMediaParseTimeout
}

// handleEndFile processes the EventEnd event during file loading
func (m *mpvPlayer) handleEndFile(event *mpv.Event) error {

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "EventEnd received during loading")

	endFile := event.EndFile()

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("EndFile reason: %s, error: %v", endFile.Reason.String(), endFile.Error))

	var validationErr error

	if endFile.Error != nil {
		validationErr = fmt.Errorf("failed to load file: %w", endFile.Error)
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

// checkVideoDimensions validates that the video has valid dimensions
func (m *mpvPlayer) checkVideoDimensions() error {

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "checkVideoDimensions: starting dimension check")

	// Get video width
	width, err := m.player.GetProperty("width", mpv.FormatInt64)
	if err != nil {
		return fmt.Errorf(errFormat, "failed to get video width", err)
	}

	// Get video height
	height, err := m.player.GetProperty("height", mpv.FormatInt64)
	if err != nil {
		return fmt.Errorf(errFormat, "failed to get video height", err)
	}

	// Validate dimensions: width and height must both be non-zero values for a valid video
	widthInt, ok := width.(int64)
	if !ok {
		return errInvalidVideoDimensions
	}

	heightInt, ok := height.(int64)
	if !ok {
		return errInvalidVideoDimensions
	}

	if widthInt == 0 || heightInt == 0 {
		return errInvalidVideoDimensions
	}

	return nil
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

	return queryGuarded[int64](&m.mu, func() bool { return m.player == nil }, func() (int64, error) {

		// Check if player is initialized
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

		// scale video window size
		scaleValue := int(windowSize * 100)

		// Not going fullscreen, so set window size
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

	res, _ := queryGuarded[*playerEvent](&m.mu, func() bool { return m.player == nil }, func() (*playerEvent, error) {

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

	m.mu.Lock() // create write lock
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

		m.player = nil // Handle is now destroyed; future RLock calls will fail gracefully

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "destroyed MPV handle: C resources released")
	}
}

// drainEvents drains any remaining events from MPV's event queue, preventing mpv from blocking
// on unprocessed events during shutdown
func (m *mpvPlayer) drainEvents() {

	drained := 0

	// Drain events for up to one second
	for range 10 {

		event := m.player.WaitEvent(0.1)
		if event == nil {
			break
		}

		drained++
	}

}
