package video

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// VLC-specific error definitions
var (
	errNoVideoTrack = errors.New("media file does not contain a video track")
)

// vlcPlayer is a wrapper around the go-vlc client
type vlcPlayer struct {
	player       *vlc.Player
	marquee      *vlc.Marquee
	eventChan    chan eventID
	screenWidth  int
	screenHeight int
	videoWidth   int
	videoHeight  int
}

// newVLCPlayer creates a new vlcPlayer instance
func newVLCPlayer() (*vlcPlayer, error) {

	// Initialize VLC library
	if err := vlc.Init("--no-video-title-show", "--quiet"); err != nil {
		return nil, fmt.Errorf(errFormat, "failed to initialize VLC library", err)
	}

	// Create player
	player, err := vlc.NewPlayer()
	if err != nil {

		if releaseErr := vlc.Release(); releaseErr != nil {
			logger.Error(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release VLC library: %v", releaseErr))
		}

		return nil, fmt.Errorf(errFormat, "failed to create VLC player", err)
	}

	// Get display resolution (needed by VLC for accurate video playback scaling)
	var displayWidth int
	var displayHeight int

	if displayWidth, displayHeight, err = screenResolution(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to get screen resolution: %v", err))
	}

	return &vlcPlayer{
		player:       player,
		marquee:      player.Marquee(),
		eventChan:    make(chan eventID, 1),
		screenWidth:  displayWidth,
		screenHeight: displayHeight,
	}, nil

}

// loadFile loads a video file into the VLC player
func (v *vlcPlayer) loadFile(path string) error {

	media, err := v.player.LoadMediaFromPath(path)
	if err != nil {
		return fmt.Errorf(errFormat, errFailedToLoadVideo, err)
	}

	defer func() {
		if err := media.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release VLC media: %v", err))
		}
	}()

	// Parse the media to validate format and metadata
	if err := v.parseMedia(media); err != nil {
		return err
	}

	// Validate the media content (dimensions)
	if err := v.validateMedia(media); err != nil {
		return err
	}

	return wrapError("failed to play video", v.player.Play())
}

// parseMedia handles the asynchronous parsing of the media file
func (v *vlcPlayer) parseMedia(media *vlc.Media) error {

	manager, err := media.EventManager()
	if err != nil {
		return fmt.Errorf(errFormat, "failed to get media event manager", err)
	}

	parseDone := make(chan struct{})

	eventCallback := func(_ vlc.Event, _ any) {
		close(parseDone)
	}

	eventID, err := manager.Attach(vlc.MediaParsedChanged, eventCallback, nil)
	if err != nil {
		return fmt.Errorf(errFormat, "failed to attach event handler", err)
	}

	defer manager.Detach(eventID)

	// Initiate parsing (async), with 5s timeout
	parseTimeout := 5
	if err := media.ParseWithOptions(parseTimeout*1000, vlc.MediaParseLocal); err != nil {
		return fmt.Errorf(errFormat, "failed to initiate media parsing", err)
	}

	// Wait here for parse completion or timeout
	select {
	case <-parseDone:
		return nil
	case <-time.After(time.Duration(parseTimeout) * time.Second):
		return errMediaParseTimeout
	}

}

// validateMedia checks the parsed media for validity (video track with dimensions)
func (v *vlcPlayer) validateMedia(media *vlc.Media) error {

	// Get tracks to validate video presence and dimensions
	tracks, err := media.Tracks()
	if err != nil {
		return fmt.Errorf(errFormat, "failed to retrieve video tracks", err)
	}

	// Look for a video track with valid dimensions
	for _, track := range tracks {

		if track.Type == vlc.MediaTrackVideo {

			// Move video dimension fields into local vars (avoid potential uint-->int conversion issues)
			trackWidth := track.Video.Width
			trackHeight := track.Video.Height

			// Check video dimensions (width and height must both be non-zero, else an invalid video file format)
			if trackWidth == 0 || trackHeight == 0 {
				return errInvalidVideoDimensions
			}

			// Ensure dimensions will not overflow during uint-->int conversion
			if trackWidth > uint(math.MaxInt) || trackHeight > uint(math.MaxInt) {
				return errInvalidVideoDimensions
			}

			// Safely convert the local variables (look Ma, no linter warns!)
			v.videoWidth = int(trackWidth)
			v.videoHeight = int(trackHeight)

			logger.Debug(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("found native video dimensions: %dx%d", trackWidth, trackHeight))

			return nil
		}

	}

	return errNoVideoTrack
}

// setSpeed sets the playback speed of the video
func (v *vlcPlayer) setSpeed(speed float64) error {
	return wrapError("failed to set video playback speed", v.player.SetPlaybackRate(float32(speed)))
}

// setPause sets the pause state of the video
func (v *vlcPlayer) setPause(paused bool) error {
	return wrapError("failed to pause video", v.player.SetPause(paused))
}

// timeRemaining gets the remaining time of the video
func (v *vlcPlayer) timeRemaining() (int64, error) {

	// Check if player is initialized
	if v.player == nil {
		return 0, errPlayerNotInitialized
	}

	length, err := v.player.MediaLength()
	if err != nil {
		return 0, fmt.Errorf(errFormat, errTimeRemaining, err)
	}

	currentTime, err := v.player.MediaTime()
	if err != nil {
		return 0, fmt.Errorf(errFormat, errTimeRemaining, err)
	}

	return (int64)((length - currentTime) / 1000), nil // Convert ms to seconds
}

// setPlaybackSize sets media player window size
func (v *vlcPlayer) setPlaybackSize(windowSize float64) error {

	// Check if screen dimensions are valid
	invalidScreenSize := v.screenWidth == 0 || v.screenHeight == 0

	// If screen dimensions are invalid, then we can only set fullscreen
	if invalidScreenSize {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, "invalid screen dimensions; will attempt to set fullscreen...")
	}

	// Enable fullscreen if window size is 1.0 (100%)
	if windowSize == 1.0 || invalidScreenSize {
		return wrapError("failed to enable fullscreen", v.player.SetFullScreen(true))
	}

	// Scale video window size based on video resolution relative to screen dimensions
	if v.videoHeight > v.videoWidth {
		windowSize *= (float64(v.screenHeight) / float64(v.videoHeight))
	} else {
		windowSize *= (float64(v.screenWidth) / float64(v.videoWidth))
	}

	// Not going fullscreen, so set window size
	return wrapError("failed to set window size", v.player.SetScale(windowSize))
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

	return wrapError(errUnableToSeek.Error(), v.player.SetMediaTime(timeMs))
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

	if _, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil); err != nil {
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

	if v.player == nil || v.marquee == nil {
		return errPlayerNotInitialized
	}

	return wrapError("failed to set marquee text", v.marquee.SetText(text))
}

// terminatePlayer cleans up VLC resources
func (v *vlcPlayer) terminatePlayer() {

	if v.eventChan != nil {
		close(v.eventChan)
		v.eventChan = nil
	}

	if v.player != nil {

		if err := v.player.Stop(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to stop VLC player: %v", err))
		}

		if err := v.player.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release VLC player: %v", err))
		}

		v.player = nil

	}

	if err := vlc.Release(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release VLC library: %v", err))
	}

}
