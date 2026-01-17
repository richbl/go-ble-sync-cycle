package video

import (
	"errors"
	"fmt"
)

// Error definitions
var (
	ErrVideoComplete = errors.New("video playback completed")

	errInvalidTimeFormat         = errors.New("invalid time format")
	errOSDUpdate                 = errors.New("failed to update OSD")
	errUnsupportedVideoPlayer    = errors.New("unsupported video player specified")
	errPlayerNotInitialized      = errors.New("media player not initialized")
	errMediaParseTimeout         = errors.New("timeout waiting for media parsing")
	errInvalidVideoDimensions    = errors.New("video dimensions are invalid")
	errPlaybackEndedUnexpectedly = errors.New("playback ended unexpectedly")
)

const (
	errFormat        = "%v: %w"
	errTimeRemaining = "failed to get time remaining in video"
	setSpeed         = "setSpeed"
	setPause         = "setPause"
)

// PlayerEventID represents the type of a player event
type eventID int

const (
	eventNone    eventID = iota // No event occurred (e.g., on timeout)
	eventEndFile                // The end of the video file has been reached
)

// playerEvent is a generic struct for player events
type playerEvent struct {
	id eventID
}

// osdConfig manages the configuration for the On-Screen Display (OSD)
type osdConfig struct {
	fontSize             int
	marginX              int
	marginY              int
	showOSD              bool
	displayCycleSpeed    bool
	displayPlaybackSpeed bool
	displayTimeRemaining bool
}

// mediaPlayer defines the interface abstraction for a video player
type mediaPlayer interface {

	// Playback methods
	loadFile(path string) error
	setSpeed(speed float64) error
	setPause(paused bool) error
	timeRemaining() (int64, error)
	terminatePlayer()

	// Configuration methods
	setFullscreen(fullscreen bool) error
	setKeepOpen(keepOpen bool) error // Used by mpv to prevent application exit on video EOF
	seek(position string) error
	setOSD(options osdConfig) error

	// Event handling methods
	setupEvents() error
	waitEvent(timeout float64) *playerEvent

	// On Screen Display (OSD) methods
	showOSDText(text string) error
}

// wrapError helper function adds return context only if an error occurred
func wrapError(context string, err error) error {

	if err == nil {
		return nil
	}

	return fmt.Errorf(errFormat, context, err)
}
