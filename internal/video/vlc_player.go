package video

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// VLC-specific error definitions
var (
	errMediaErrorState       = errors.New("MediaError state reported")
	errSeekTimeOverflow      = errors.New("seek time out of bounds")
	errVLCStreamNotSeekable  = errors.New("timeout waiting for VLC stream to become seekable")
	errVLCVoutNotInitialized = errors.New("timeout waiting for VLC vout initialization")
)

const (
	errFailedToReleaseVLCLibrary       = "failed to release VLC library"
	errFailedToReleaseVLCPlayer        = "failed to release VLC player"
	errFailedToReleaseValidationPlayer = "failed to release validation player"
	errFailedToReleaseValidationMedia  = "failed to release validation media"
)

// vlcPlayer is a wrapper around the go-vlc client
type vlcPlayer struct {
	player    *vlc.Player
	mu        sync.RWMutex
	marquee   *vlc.Marquee
	eventChan chan playerEvent

	// Video playback properties
	videoWidth        int
	videoHeight       int
	duration          float64
	pendingWindowSize float64
}

// newVLCPlayer creates a new vlcPlayer instance
func newVLCPlayer(ctx context.Context) (*vlcPlayer, error) {

	// Initialize for playback (not headless)
	if err := initVLCLibrary(false); err != nil {
		return nil, err
	}

	player, err := vlc.NewPlayer()
	if err != nil {

		if relErr := vlc.Release(); relErr != nil {
			logger.Warn(ctx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseVLCLibrary+": %v", relErr))
		}

		return nil, fmt.Errorf("failed to create VLC player: %w", err)
	}

	logger.Info(ctx, logger.VIDEO, "VLC player object created")

	return &vlcPlayer{
		player:    player,
		marquee:   player.Marquee(),
		eventChan: make(chan playerEvent, 1),
	}, nil
}

// initVLCLibrary initializes VLC with appropriate flags
func initVLCLibrary(headless bool) error {

	var args []string

	// Set appropriate flags
	if headless {
		args = []string{"--quiet", "--no-xlib", "--vout=dummy", "--aout=dummy"}
	} else {
		args = []string{"--no-video-title-show", "--quiet"}
	}

	if err := vlc.Init(args...); err != nil {
		return fmt.Errorf("failed to initialize VLC library: %w", err)
	}

	return nil
}

// loadFile loads a video file into the VLC player and validates it
func (v *vlcPlayer) loadFile(path string) error {

	v.mu.Lock()
	defer v.mu.Unlock()

	// Reset existing playback instance
	v.resetPlayer()

	logger.Debug(logger.BackgroundCtx, logger.VIDEO, "starting headless video validation...")

	metadata, err := v.validateVideoFile(path)
	if err != nil {

		// Attempt to restore normal VLC state even when validation fails
		_ = initVLCLibrary(false)

		return fmt.Errorf(errFormat, errFailedToLoadVideo, err)
	}

	logger.Info(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("video validation successful: %dx%d, %.2fs", metadata.width, metadata.height, metadata.duration))

	// Re-initialize player for actual playback
	if err := v.initializePlaybackPlayer(metadata); err != nil {
		return err
	}

	// Load media for playback
	media, err := v.player.LoadMediaFromPath(path)
	if err != nil {
		return fmt.Errorf(errFormat, errFailedToLoadVideo, err)
	}

	defer func() {

		if rErr := media.Release(); rErr != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release media: %v", rErr))
		}

	}()

	if err := v.player.SetMedia(media); err != nil {
		return fmt.Errorf("failed to set media: %w", err)
	}

	// Apply staged window size pre-Play() so VLC uses it during vout creation,
	// sizing the window itself rather than only scaling content within it
	if v.pendingWindowSize > 0 {
		if err := v.applyWindowSize(v.pendingWindowSize); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("pre-play window size failed: %v", err))
		}
	}

	return v.startPlaybackAndWaitForVout()
}

// initializePlaybackPlayer initializes the player for playback with validated metadata
func (v *vlcPlayer) initializePlaybackPlayer(metadata *videoValidationInfo) error {

	if err := initVLCLibrary(false); err != nil {
		return err
	}

	player, err := vlc.NewPlayer()
	if err != nil {

		if relErr := vlc.Release(); relErr != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseVLCLibrary+": %v", relErr))
		}

		return fmt.Errorf("failed to recreate VLC player: %w", err)
	}

	v.player = player

	// Cache marquee object for OSD display
	v.marquee = v.player.Marquee()

	// Cache validated metadata
	v.videoWidth = metadata.width
	v.videoHeight = metadata.height
	v.duration = metadata.duration

	return nil
}

// startPlaybackAndWaitForVout starts playback and waits for the vout to be initialized
func (v *vlcPlayer) startPlaybackAndWaitForVout() error {

	// Subscribe to MediaPlayerVout BEFORE Play() so the event cannot be missed
	manager, err := v.player.EventManager()
	if err != nil {
		return fmt.Errorf("failed to get event manager: %w", err)
	}

	voutReady := make(chan struct{}, 1)

	voutEventID, err := manager.Attach(vlc.MediaPlayerVout, func(_ vlc.Event, _ any) {

		select {
		case voutReady <- struct{}{}:
		default:
		}

	}, nil)
	if err != nil {
		return fmt.Errorf("failed to attach MediaPlayerVout event: %w", err)
	}

	defer manager.Detach(voutEventID)

	if err := v.player.Play(); err != nil {
		return wrapError("failed to play video", err)
	}

	select {
	case <-voutReady:
		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "VLC vout confirmed via MediaPlayerVout event")

	case <-time.After(10 * time.Second):
		return errVLCVoutNotInitialized
	}

	return nil
}

// validateVideoFile performs headless validation of a video file
func (v *vlcPlayer) validateVideoFile(path string) (*videoValidationInfo, error) {

	if err := initVLCLibrary(true); err != nil {
		return nil, fmt.Errorf("validation init failed: %w", err)
	}

	defer v.releaseValidationLibrary()

	// Create and setup validation player
	player, media, err := v.createValidationPlayer(path)
	if err != nil {
		return nil, err
	}

	defer v.releaseValidationPlayer(player, media)

	// Start playback to force decoder initialization
	if err := player.Play(); err != nil {
		return nil, fmt.Errorf("validation playback failed: %w", err)
	}

	// Wait for decoder to populate track metadata
	metadata, err := v.pollForActiveStream(player, media)
	if err != nil {
		return nil, err
	}

	// Retrieve duration
	v.extractDuration(player, media, metadata)

	// Stop validation playback
	if err := player.Stop(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("validation player stop failed: %v", err))
	}

	return metadata, nil
}

// releaseValidationLibrary releases the VLC library used for validation
func (v *vlcPlayer) releaseValidationLibrary() {

	if err := vlc.Release(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release validation library: %v", err))
	}

}

// createValidationPlayer creates a player and loads media for validation
func (v *vlcPlayer) createValidationPlayer(path string) (*vlc.Player, *vlc.Media, error) {

	player, err := vlc.NewPlayer()
	if err != nil {
		return nil, nil, fmt.Errorf("validation player creation failed: %w", err)
	}

	media, err := vlc.NewMediaFromPath(path)
	if err != nil {

		if relErr := player.Release(); relErr != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseValidationPlayer+": %v", relErr))
		}

		return nil, nil, fmt.Errorf("validation media load failed: %w", err)
	}

	if err := player.SetMedia(media); err != nil {

		if relErr := media.Release(); relErr != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseValidationMedia+": %v", relErr))
		}

		if relErr := player.Release(); relErr != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseValidationPlayer+": %v", relErr))
		}

		return nil, nil, fmt.Errorf("validation SetMedia failed: %w", err)
	}

	return player, media, nil
}

// releaseValidationPlayer releases validation player and media resources
func (v *vlcPlayer) releaseValidationPlayer(player *vlc.Player, media *vlc.Media) {

	if media != nil {

		if err := media.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseValidationMedia+": %v", err))
		}

	}

	if player != nil {

		if err := player.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseValidationPlayer+": %v", err))
		}

	}

}

// extractDuration retrieves duration from player or media
func (v *vlcPlayer) extractDuration(player *vlc.Player, media *vlc.Media, metadata *videoValidationInfo) {

	if length, err := player.MediaLength(); err == nil {
		metadata.duration = float64(length) / 1000.0
	} else if dur, err := media.Duration(); err == nil {
		metadata.duration = float64(dur) / 1000.0
	}

}

// pollForActiveStream waits for video codec to appear and extracts stream information
func (v *vlcPlayer) pollForActiveStream(player *vlc.Player, media *vlc.Media) (*videoValidationInfo, error) {

	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, errStreamTimeout

		case <-ticker.C:
			metadata, done, err := v.extractStreamInfo(player, media)
			if err != nil {
				return nil, err
			}

			if done {
				return metadata, nil
			}

		}
	}

}

// extractStreamInfo examines the current player and media state to determine if valid metadata is available
func (v *vlcPlayer) extractStreamInfo(player *vlc.Player, media *vlc.Media) (*videoValidationInfo, bool, error) {

	state, err := player.MediaState()
	if err != nil {
		return nil, true, fmt.Errorf("failed to get media state: %w", err)
	}

	if state == vlc.MediaError {
		return nil, true, errMediaErrorState
	}

	// Attempt to get track information
	tracks, err := media.Tracks()
	if err != nil {
		return nil, true, fmt.Errorf(errFormat, errNoVideoTrack, err)
	}

	if len(tracks) > 0 {
		metadata, found := getTrackInfo(tracks)
		if found {

			// Validate we got real dimensions (not 0x0)
			if metadata.width == 0 || metadata.height == 0 {

				// Video track exists but dimensions not yet populated, so keep waiting
				return nil, false, nil
			}

			return metadata, true, nil
		}
	}

	// Playback ended without finding valid video track
	if state == vlc.MediaEnded {
		return nil, true, errNoVideoTrack
	}

	return nil, false, nil
}

// getTrackInfo searches tracks for video track and extracts dimensions
func getTrackInfo(tracks []*vlc.MediaTrack) (*videoValidationInfo, bool) {

	for _, track := range tracks {

		if track.Type == vlc.MediaTrackVideo && track.Video != nil {

			// Move video dimension fields into local vars (avoid potential uint-->int conversion issues)
			trackWidth := track.Video.Width
			trackHeight := track.Video.Height

			// Check bounds before conversion
			if trackWidth > uint(math.MaxInt) || trackHeight > uint(math.MaxInt) {
				logger.Warn(logger.BackgroundCtx, logger.VIDEO, "video dimensions exceed integer limits")

				return nil, false
			}

			return &videoValidationInfo{
				width:  int(trackWidth),
				height: int(trackHeight),
			}, true
		}

	}

	return nil, false
}

// setSpeed sets the playback speed of the video
func (v *vlcPlayer) setSpeed(speed float64) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil }, func() error {
		return wrapError("failed to set video playback speed", v.player.SetPlaybackRate(float32(speed)))
	})
}

// setPause sets the pause state of the video
func (v *vlcPlayer) setPause(paused bool) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil }, func() error {
		return wrapError("failed to pause video", v.player.SetPause(paused))
	})
}

// timeRemaining gets the remaining time of the video
func (v *vlcPlayer) timeRemaining() (int64, error) {

	return queryGuarded(&v.mu, func() bool { return v.player == nil }, func() (int64, error) {

		currentTime, err := v.player.MediaTime()
		if err != nil {
			return 0, fmt.Errorf("failed to get current media time: %w", err)
		}

		// v.duration is cached from validateVideoFile()
		mediaLength := int(v.duration * 1000)

		return int64((mediaLength - currentTime) / 1000), nil
	})
}

// setPlaybackSize sets the playback size based on the provided window size percentage
func (v *vlcPlayer) setPlaybackSize(windowSize float64) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil }, func() error {

		v.pendingWindowSize = windowSize

		// Dimensions only known after validateVideoFile()
		if v.videoWidth == 0 || v.videoHeight == 0 {
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "window size staged for pre-Play() application in loadFile()")

			return nil
		}

		return v.applyWindowSize(windowSize)
	})
}

// applyWindowSize applies the window size scaling based on video and display dimensions
func (v *vlcPlayer) applyWindowSize(windowSize float64) error {

	displayWidth, displayHeight, err := screenResolution()
	if err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to acquire screen resolution: %v", err))
	}

	// Validate display dimensions
	if displayWidth == 0 || displayHeight == 0 {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, "invalid display dimensions; defaulting to fullscreen")

		return wrapError("failed to enable fullscreen", v.player.SetFullScreen(true))
	}

	// Fullscreen if window size is 100%
	if windowSize == 1.0 {
		return wrapError("failed to enable fullscreen", v.player.SetFullScreen(true))
	}

	// Scale window size based on video aspect ratio relative to display
	if v.videoHeight > 0 && v.videoWidth > 0 {

		if v.videoHeight > v.videoWidth {

			// Portrait video: scale based on display height
			windowSize *= (float64(displayHeight) / float64(v.videoHeight))
		} else {

			// Landscape video: scale based on display width
			windowSize *= (float64(displayWidth) / float64(v.videoWidth))
		}

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "video playback size successfully set")
	}

	return wrapError("failed to set window size", v.player.SetScale(windowSize))
}

// Stub: setKeepOpen is not supported in VLC
func (v *vlcPlayer) setKeepOpen(_ bool) error {
	return nil
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

	// Bounds check to prevent overflow
	if totalSeconds > int64(math.MaxInt/1000) {
		return 0, errSeekTimeOverflow
	}

	return int(totalSeconds * 1000), nil
}

// parseMMSS parses "MM:SS" format and returns total seconds
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

// parseSS parses "SS" format and returns total seconds
func parseSS(position string) (int64, error) {

	totalSeconds, err := strconv.ParseInt(position, 10, 64)
	if err != nil || totalSeconds < 0 {
		return 0, fmt.Errorf(errFormat, position, errInvalidTimeFormat)
	}

	return totalSeconds, nil
}

// seek seeks to a specific position in the video based on a time string
func (v *vlcPlayer) seek(position string) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil }, func() error {

		timeMs, err := parseTimePosition(position)
		if err != nil {
			return fmt.Errorf("unable to parse specified seek time: %w", err)
		}

		manager, err := v.player.EventManager()
		if err != nil {
			return fmt.Errorf("failed to get event manager for seek: %w", err)
		}

		seekable := make(chan struct{}, 1)

		// Check if already seekable before subscribing to avoid unnecessary wait
		if v.player.IsSeekable() {
			close(seekable)
		} else {
			eventID, err := manager.Attach(vlc.MediaPlayerSeekableChanged, func(_ vlc.Event, _ any) {

				select {
				case seekable <- struct{}{}:
				default:
				}

			}, nil)

			if err != nil {
				return fmt.Errorf("failed to attach seekable event: %w", err)
			}
			defer manager.Detach(eventID)
		}

		select {
		case <-seekable:
			logger.Debug(logger.BackgroundCtx, logger.VIDEO, "VLC stream seekable confirmed")

		case <-time.After(5 * time.Second):
			return errVLCStreamNotSeekable
		}

		return wrapError("unable to seek", v.player.SetMediaTime(timeMs))
	})
}

// setOSD configures the On-Screen Display (OSD)
func (v *vlcPlayer) setOSD(options osdConfig) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil || v.marquee == nil }, func() error {

		if err := v.marquee.SetX(options.marginX); err != nil {
			return fmt.Errorf("failed to set OSD X position: %w", err)
		}

		if err := v.marquee.SetY(options.marginY); err != nil {
			return fmt.Errorf("failed to set OSD Y position: %w", err)
		}

		if err := v.marquee.SetSize(options.fontSize); err != nil {
			return fmt.Errorf("failed to set OSD font size: %w", err)
		}

		if err := v.marquee.Enable(true); err != nil {
			return fmt.Errorf("failed to enable OSD: %w", err)
		}

		return nil
	})
}

// setupEvents subscribes to VLC playback events
func (v *vlcPlayer) setupEvents() error {

	return execGuarded(&v.mu, func() bool { return v.player == nil }, func() error {

		manager, err := v.player.EventManager()
		if err != nil {
			return fmt.Errorf("failed to get VLC event manager: %w", err)
		}

		// eventCallback is triggered when video playback ends
		eventCallback := func(_ vlc.Event, _ any) {

			v.mu.RLock()
			defer v.mu.RUnlock()

			if v.eventChan == nil {
				return
			}

			// Use non-blocking send to prevent VLC thread hangs if channel is full
			select {
			case v.eventChan <- playerEvent{id: eventEndFile}:
			default:
				logger.Warn(context.Background(), logger.VIDEO, "event channel full; dropped VLC event")
			}
		}

		if _, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil); err != nil {
			return fmt.Errorf("failed to attach VLC event handler: %w", err)
		}

		return nil
	})
}

// waitEvent waits for a player event with timeout
func (v *vlcPlayer) waitEvent(timeout float64) *playerEvent {

	select {
	case event, ok := <-v.eventChan:
		if !ok {
			return &playerEvent{id: eventNone}
		}

		return &event

	case <-time.After(time.Duration(timeout * float64(time.Second))):
		return &playerEvent{id: eventNone}
	}
}

// showOSDText displays text on the video using VLC's marquee feature
func (v *vlcPlayer) showOSDText(text string) error {

	return execGuarded(&v.mu, func() bool { return v.player == nil || v.marquee == nil }, func() error {
		return wrapError("failed to set marquee text", v.marquee.SetText(text))
	})
}

// resetPlayer releases player and library resources
func (v *vlcPlayer) resetPlayer() {

	if v.player != nil {

		if err := v.player.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseVLCPlayer+": %v", err))
		}

		v.player = nil
		v.marquee = nil
	}

	if err := vlc.Release(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to release library: %v", err))
	}

}

// terminatePlayer cleans up VLC resources
func (v *vlcPlayer) terminatePlayer() {

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.player != nil {
		if err := v.player.Stop(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf("failed to stop VLC player: %v", err))
		}

		if err := v.player.Release(); err != nil {
			logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseVLCPlayer+": %v", err))
		}

		v.player = nil
		v.marquee = nil

		logger.Debug(logger.BackgroundCtx, logger.VIDEO, "VLC player stopped and instance released")
	}

	if v.eventChan != nil {
		close(v.eventChan)
		v.eventChan = nil
	}

	if err := vlc.Release(); err != nil {
		logger.Warn(logger.BackgroundCtx, logger.VIDEO, fmt.Sprintf(errFailedToReleaseVLCPlayer+": %v", err))
	}

}
