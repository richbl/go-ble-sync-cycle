package video

import (
	"errors"
)

// Error definitions
var (
	errInvalidTimeFormat = errors.New("invalid time format")
	errOSDUpdate         = errors.New("failed to update OSD")
	errVideoComplete     = errors.New("video playback completed")
)

const (
	errFormat        = "%v: %w"
	errTimeRemaining = "failed to get time remaining in video"
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
	showOSD              bool
	fontSize             int
	displayCycleSpeed    bool
	displayPlaybackSpeed bool
	displayTimeRemaining bool
	marginX              int
	marginY              int
}

// mediaPlayer defines the interface abstraction for a video player
type mediaPlayer interface {
	loadFile(path string) error
	setSpeed(speed float64) error
	setPause(paused bool) error
	getTimeRemaining() (int64, error)

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
	terminatePlayer()
}
