package video

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// vlcPlayer is a wrapper around the go-vlc client
type vlcPlayer struct {
	player    *vlc.Player
	marquee   *vlc.Marquee
	eventChan chan eventID
}

// newVLCPlayer creates a new vlcPlayer instance
func newVLCPlayer() (*vlcPlayer, error) {

	// Initialize VLC library
	if err := vlc.Init("--no-video-title-show", "--quiet"); err != nil {
		return nil, err
	}

	// Create player
	player, err := vlc.NewPlayer()
	if err != nil {

		if releaseErr := vlc.Release(); releaseErr != nil {
			logger.Error(logger.VIDEO, fmt.Sprintf("failed to release VLC library: %v", releaseErr))
		}

		return nil, err
	}

	return &vlcPlayer{
		player:    player,
		marquee:   player.Marquee(),
		eventChan: make(chan eventID, 1),
	}, nil

}

// loadFile loads a video file into the VLC player
func (v *vlcPlayer) loadFile(path string) error {

	media, err := v.player.LoadMediaFromPath(path)
	if err != nil {
		return err
	}

	defer func() {
		if err := media.Release(); err != nil {
			logger.Warn(logger.VIDEO, fmt.Sprintf("failed to release VLC media: %v", err))
		}
	}()

	return v.player.Play()
}

// setSpeed sets the playback speed of the video
func (v *vlcPlayer) setSpeed(speed float64) error {
	return v.player.SetPlaybackRate(float32(speed))
}

// setPause sets the pause state of the video
func (v *vlcPlayer) setPause(paused bool) error {
	return v.player.SetPause(paused)
}

// getTimeRemaining gets the remaining time of the video
func (v *vlcPlayer) getTimeRemaining() (int64, error) {

	length, err := v.player.MediaLength()
	if err != nil {
		return 0, fmt.Errorf(errFormat, errTimeRemaining, err)
	}

	currentTime, err := v.player.MediaTime()
	if err != nil {
		return 0, err
	}

	return (int64)((length - currentTime) / 1000), nil // Convert ms to seconds
}

// setFullscreen toggles fullscreen mode
func (v *vlcPlayer) setFullscreen(fullscreen bool) error {

	if err := v.player.SetFullScreen(fullscreen); err != nil {
		return fmt.Errorf(errFormat, "failed to set fullscreen", err)
	}

	return nil
}

// Stub: setKeepOpen is not supported in VLC
func (v *vlcPlayer) setKeepOpen(_ bool) error {
	return nil
}

// parseTimePosition parses a time string in "MM:SS" or "SS" format and converts it to milliseconds
func parseTimePosition(position string) (int, error) {

	position = strings.TrimSpace(position)
	var totalSeconds int64
	var err error

	if strings.Contains(position, ":") {
		totalSeconds, err = parseMMSS(position)
	} else {
		totalSeconds, err = parseSS(position)
	}

	if err != nil {
		return 0, err
	}

	return int(totalSeconds * 1000), nil
}

// parseMMSS parses a time string in "MM:SS" format and returns total seconds
func parseMMSS(position string) (int64, error) {

	parseErr := fmt.Errorf("%s is an %w", position, errInvalidTimeFormat)

	parts := strings.Split(position, ":")
	if len(parts) != 2 {
		return 0, parseErr
	}

	minutes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || minutes < 0 {
		return 0, parseErr
	}

	seconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || seconds < 0 || seconds > 59 {
		return 0, parseErr
	}

	return (minutes * 60) + seconds, nil
}

// parseSS parses a time string in "SS" format and returns total seconds
func parseSS(position string) (int64, error) {

	totalSeconds, err := strconv.ParseInt(position, 10, 64)
	if err != nil || totalSeconds < 0 {
		return 0, fmt.Errorf("%s is an %w", position, errInvalidTimeFormat)
	}

	return totalSeconds, nil
}

// seek moves the playback position to the specified time (MM:SS format)
func (v *vlcPlayer) seek(position string) error {

	timeMs, err := parseTimePosition(position)
	if err != nil {
		return fmt.Errorf(errFormat, "unable to parse specified seek time", err)
	}

	return v.player.SetMediaTime(timeMs)
}

// setOSD configures the On-Screen Display (OSD)
func (v *vlcPlayer) setOSD(options osdConfig) error {

	if err := v.marquee.SetX(options.marginX); err != nil {
		return fmt.Errorf(errFormat, "failed to set OSD X position", err)
	}

	if err := v.marquee.SetY(options.marginY); err != nil {
		return fmt.Errorf(errFormat, "failed to set OSD Y position", err)
	}

	if err := v.marquee.SetSize(options.fontSize); err != nil {
		return fmt.Errorf(errFormat, "failed to set OSD font size", err)
	}

	if err := v.marquee.Enable(true); err != nil {
		return fmt.Errorf(errFormat, "failed to enable OSD", err)
	}

	return nil
}

// setupEvents subscribes to VLC playback events
func (v *vlcPlayer) setupEvents() error {

	manager, err := v.player.EventManager()
	if err != nil {
		return fmt.Errorf(errFormat, "failed to get VLC event manager", err)
	}

	// eventCallback is triggered when the video playback ends
	eventCallback := func(_ vlc.Event, _ any) {
		v.eventChan <- eventEndFile
	}

	_, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		return fmt.Errorf(errFormat, "failed to attach VLC event handler", err)
	}

	return nil
}

// waitEvent waits for a player event with timeout
func (v *vlcPlayer) waitEvent(timeout float64) *playerEvent {

	select {
	case eventID := <-v.eventChan:
		return &playerEvent{id: eventID}
	case <-time.After(time.Duration(timeout * float64(time.Second))):
		return &playerEvent{id: eventNone}
	}

}

// showOSDText displays text on the video using VLC's marquee feature
func (v *vlcPlayer) showOSDText(text string) error {

	if err := v.marquee.SetText(text); err != nil {
		return fmt.Errorf(errFormat, "failed to set marquee text", err)
	}

	return nil
}

// terminatePlayer cleans up VLC resources
func (v *vlcPlayer) terminatePlayer() {

	if v.eventChan != nil {
		close(v.eventChan)
		v.eventChan = nil
	}

	if v.player != nil {

		if err := v.player.Stop(); err != nil {
			logger.Warn(logger.VIDEO, fmt.Sprintf("failed to stop VLC player: %v", err))
		}

		if err := v.player.Release(); err != nil {
			logger.Warn(logger.VIDEO, fmt.Sprintf("failed to release VLC player: %v", err))
		}

		v.player = nil

	}

	if err := vlc.Release(); err != nil {
		logger.Warn(logger.VIDEO, fmt.Sprintf("failed to release VLC library: %v", err))
	}

}
