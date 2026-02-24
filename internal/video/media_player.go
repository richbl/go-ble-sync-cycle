package video

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
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
	errNoVideoTrack              = errors.New("media file does not contain a video track")
	errFailedToCreatePlayer      = errors.New("failed to instantiate media player")
	errStreamTimeout             = errors.New("timeout waiting for video stream")
	errPlaybackEndedUnexpectedly = errors.New("playback ended unexpectedly")
	errFailedToLoadVideo         = errors.New("failed to load video")
	errUnableToSeek              = errors.New("failed to seek to specified position in media player")
	ErrVideoComplete             = errors.New("video playback completed")

	// GLFW-specific errors
	errFailedToInitializeGLFW = errors.New("failed to initialize GLFW")
	errFailedToAcquireMonitor = errors.New("failed to acquire primary monitor (GLFW)")
	errFailedToGetVideoMode   = errors.New("failed to get video mode (GLFW)")
)

// videoValidationInfo holds video properties discovered when validating a media file
type videoValidationInfo struct {
	width    int
	height   int
	duration float64
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

// screenResolution returns the screen resolution of the primary monitor (needed by VLC for video
// playback scaling)
func screenResolution() (int, int, error) {

	// Initialize framework
	if err := glfw.Init(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("%v: %v", errFailedToInitializeGLFW, err))

		return 0, 0, errFailedToInitializeGLFW
	}

	defer glfw.Terminate()

	// Get the primary monitor
	monitor := glfw.GetPrimaryMonitor()
	if monitor == nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, errFailedToAcquireMonitor)

		return 0, 0, errFailedToAcquireMonitor
	}

	// Get the current video dimensions (width, height)
	mode := monitor.GetVideoMode()
	if mode == nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, errFailedToGetVideoMode)

		return 0, 0, errFailedToGetVideoMode
	}

	return mode.Width, mode.Height, nil
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
