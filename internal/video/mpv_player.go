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

	mpv "github.com/gen2brain/go-mpv"
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

// loadFile loads a video file into the mpv player
func (m *mpvPlayer) loadFile(path string) error {
	return wrapError("failed to load video file", m.player.Command([]string{"loadfile", path}))
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

// setupEvents prepares the player to listen for the end-of-file event
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
	m.player.TerminateDestroy()
}
