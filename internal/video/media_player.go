package video

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Error definitions
var (
	// Playback-specific errors
	errInvalidTimeFormat         = errors.New("invalid time format")
	errOSDUpdate                 = errors.New("failed to update OSD")
	errUnsupportedVideoPlayer    = errors.New("unsupported video player specified")
	errPlayerNotInitialized      = errors.New("media player not initialized")
	errMediaParseTimeout         = errors.New("timeout waiting for media parsing")
	errInvalidVideoDimensions    = errors.New("video dimensions are invalid")
	errNoVideoTrack              = errors.New("video file does not contain a video track")
	errFailedToCreatePlayer      = errors.New("failed to instantiate media player")
	errStreamTimeout             = errors.New("timeout waiting for video stream")
	errPlaybackEndedUnexpectedly = errors.New("playback ended unexpectedly")
	errFailedToValidateVideo     = errors.New("failed to validate video file")
	errFailedToLoadVideo         = errors.New("failed to load video")
	errUnableToSeek              = errors.New("failed to seek to specified position in media player")
	ErrSeekExceedsDuration       = errors.New("seek position exceeds video file duration")

	//
	ErrVideoComplete = errors.New("video playback completed")
)

// videoValidationInfo holds video properties discovered when validating a media file
type videoValidationInfo struct {
	width  int
	height int
}

const (
	errFormat        = "%v: %w"
	errTimeRemaining = "failed to get time remaining in video"
	setSpeed         = "setSpeed"
	setPause         = "setPause"
)

// PlayerEventID represents the type of a player event
type eventID int

const (
	eventNone    eventID = iota // No media player event occurred
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
	validateVideoFile(videoPath string, position string) error
	loadFile(path string) error
	setSpeed(speed float64) error
	setPause(paused bool) error
	timeRemaining() (int64, error)
	terminatePlayer()

	// Configuration methods
	setPlaybackSize(windowSize float64) error
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

// execGuarded follows a lifecycle guard pattern to allow concurrent commands while the player is alive
func execGuarded(mu *sync.RWMutex, isNil func() bool, action func() error) error {

	mu.RLock()
	defer mu.RUnlock()

	if isNil() {
		return errPlayerNotInitialized
	}

	return action()
}

// queryGuarded permits concurrent commands while the player is alive
//
//nolint:ireturn // Legitimate use of generics for internal player abstraction
func queryGuarded[T any](mu *sync.RWMutex, isNil func() bool, action func() (T, error)) (T, error) {

	mu.RLock()
	defer mu.RUnlock()

	if isNil() {
		var zero T

		return zero, errPlayerNotInitialized
	}

	return action()
}

// parseTimePosition parses a time string in "MM:SS" or "SS" format and converts to milliseconds
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

// parseMMSS parses "MM:SS" time format and returns total seconds
func parseMMSS(position string) (int64, error) {

	parts := strings.Split(position, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf(errFormat, position, errInvalidTimeFormat)
	}

	minutes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || minutes < 0 {
		return 0, fmt.Errorf(errFormat, position, errInvalidTimeFormat)
	}

	seconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || seconds < 0 || seconds > 59 {
		return 0, fmt.Errorf(errFormat, position, errInvalidTimeFormat)
	}

	return (minutes * 60) + seconds, nil
}

// parseSS parses "SS" time format and returns total seconds
func parseSS(position string) (int64, error) {

	totalSeconds, err := strconv.ParseInt(position, 10, 64)
	if err != nil || totalSeconds < 0 {
		return 0, fmt.Errorf(errFormat, position, errInvalidTimeFormat)
	}

	return totalSeconds, nil
}
