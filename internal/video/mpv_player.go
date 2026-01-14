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
	"time"

	mpv "github.com/gen2brain/go-mpv"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// mpvPlayer is a wrapper around the go-mpv client
type mpvPlayer struct {
	player *mpv.Mpv
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

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "attempting to load file: "+path)

	if err := m.player.Command([]string{"loadfile", path}); err != nil {

		logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("mpv command failed: %v", err))

		return wrapError("failed to load video file", err)
	}

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "command succeeded, now validating file...")

	// Wait for file-loaded event and validate the file
	err := m.validateLoadedFile()
	if err != nil {
		logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("video file validation failed: %v", err))

		return err
	}

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "video file validation succeeded")

	return nil
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
			// Validation successful - drain any remaining events before returning
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "validation successful, draining remaining events")

			m.drainEvents()

			return nil

			// During file loading, ANY end event means the file failed to load
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
			validationErr = errVideoComplete
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
	return wrapError("failed to set video playback speed", m.player.SetProperty("speed", mpv.FormatDouble, speed))
}

// setPause sets the pause state of the video
func (m *mpvPlayer) setPause(paused bool) error {
	return wrapError("failed to pause video", m.player.SetProperty("pause", mpv.FormatFlag, paused))
}

// timeRemaining gets the remaining time of the video
func (m *mpvPlayer) timeRemaining() (int64, error) {

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
}

// setFullscreen toggles fullscreen mode
func (m *mpvPlayer) setFullscreen(fullscreen bool) error {

	context := "failed to disable fullscreen"
	value := "no"

	if fullscreen {
		context = "failed to enable fullscreen"
		value = "yes"
	}

	return wrapError(context, m.player.SetOptionString("fullscreen", value))
}

// setKeepOpen configures the player to keep the window open after playback completes
func (m *mpvPlayer) setKeepOpen(keepOpen bool) error {

	value := "no"

	if keepOpen {
		value = "yes"
	}

	return wrapError("failed to set keep-open media player option", m.player.SetOptionString("keep-open", value))
}

// seek moves the playback position to the specified time position
func (m *mpvPlayer) seek(position string) error {
	return wrapError("failed to seek to specified position in media player", m.player.SetOptionString("start", position))
}

// setOSD configures the On-Screen Display (OSD)
func (m *mpvPlayer) setOSD(options osdConfig) error {

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
}

// setupEvents prepares the player to listen for end-of-file and file-loaded events
func (m *mpvPlayer) setupEvents() error {
	return wrapError("failed to setup end-of-file observe event", m.player.ObserveProperty(0, "eof-reached", mpv.FormatFlag))
}

// waitEvent waits for an mpv event and translates it to a generic playerEvent
func (m *mpvPlayer) waitEvent(timeout float64) *playerEvent {

	event := m.player.WaitEvent(timeout)
	if event == nil {
		return &playerEvent{id: eventNone}
	}

	if event.EventID == mpv.EventPropertyChange {
		prop := event.Property()

		if prop.Data != nil && prop.Name == "eof-reached" {
			reached, ok := prop.Data.(bool)
			if ok && reached {
				return &playerEvent{id: eventEndFile}
			}
		}

	}

	return &playerEvent{id: eventNone}
}

// showOSDText displays text on the OSD
func (m *mpvPlayer) showOSDText(text string) error {
	return wrapError("failed to show OSD text", m.player.SetOptionString("osd-msg1", text))
}

// terminatePlayer terminates the mpv player instance and cleans up resources
func (m *mpvPlayer) terminatePlayer() {

	logger.Debug(logger.BackgroundCtx, "starting player termination")

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
			logger.Debug(logger.BackgroundCtx, "call to terminate mpv completed successfully")
		case <-time.After(2 * time.Second):
			logger.Warn(logger.BackgroundCtx, "call to terminate mpv timed out after 2s, continuing mpv shutdown")
		}

		m.player = nil
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
