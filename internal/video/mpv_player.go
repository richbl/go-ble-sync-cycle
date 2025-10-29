package video

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

	m := &mpvPlayer{
		player: mpv.New(),
	}

	if err := m.player.Initialize(); err != nil {
		return nil, err
	}

	return m, nil
}

// loadFile loads a video file into the mpv player
func (m *mpvPlayer) loadFile(path string) error {
	return m.player.Command([]string{"loadfile", path})
}

// setSpeed sets the playback speed of the video
func (m *mpvPlayer) setSpeed(speed float64) error {
	return m.player.SetProperty("speed", mpv.FormatDouble, speed)
}

// setPause sets the pause state of the video
func (m *mpvPlayer) setPause(paused bool) error {
	return m.player.SetProperty("pause", mpv.FormatFlag, paused)
}

// getTimeRemaining gets the remaining time of the video
func (m *mpvPlayer) getTimeRemaining() (int64, error) {

	timeRemaining, err := m.player.GetProperty("time-remaining", mpv.FormatInt64)
	if err != nil {
		return 0, err
	}

	timeRemainingInt, ok := timeRemaining.(int64)
	if !ok {
		return 0, errInvalidTimeFormat

	}

	return timeRemainingInt, nil
}

// setFullscreen toggles fullscreen mode
func (m *mpvPlayer) setFullscreen(fullscreen bool) error {

	if fullscreen {
		return m.player.SetOptionString("fullscreen", "yes")
	}

	return m.player.SetOptionString("fullscreen", "no")
}

// setKeepOpen configures the player to keep the window open after playback completes
func (m *mpvPlayer) setKeepOpen(keepOpen bool) error {

	if keepOpen {
		return m.player.SetOptionString("keep-open", "yes")
	}

	return m.player.SetOptionString("keep-open", "no")
}

// seek moves the playback position to the specified time position
func (m *mpvPlayer) seek(position string) error {
	return m.player.SetOptionString("start", position)
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
	return m.player.ObserveProperty(0, "eof-reached", mpv.FormatFlag)
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
	return m.player.SetOptionString("osd-msg1", text)
}

// terminatePlayer terminates the mpv player instance and cleans up resources
func (m *mpvPlayer) terminatePlayer() {
	m.player.TerminateDestroy()
}
