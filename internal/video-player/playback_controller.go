package video

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"

	"github.com/gen2brain/go-mpv"
)

// PlaybackController represents the MPV video player component and its configuration
type PlaybackController struct {
	config      config.VideoConfig
	speedConfig config.SpeedConfig
	player      *mpv.Mpv
}

// NewPlaybackController creates a new video player with the given configuration
func NewPlaybackController(videoConfig config.VideoConfig, speedConfig config.SpeedConfig) (*PlaybackController, error) {

	player := mpv.New()
	if err := player.Initialize(); err != nil {
		return nil, err
	}

	return &PlaybackController{
		config:      videoConfig,
		speedConfig: speedConfig,
		player:      player,
	}, nil

}

// Start configures and starts the MPV video player
func (p *PlaybackController) Start(ctx context.Context, speedController *speed.SpeedController) error {

	// Defer player cleanup
	defer p.player.TerminateDestroy()

	logger.Info("[VIDEO] Starting MPV video player...")

	// Scale MPV playback window size
	if err := p.player.SetOption("window-scale", mpv.FormatDouble, p.config.WindowScaleFactor); err != nil {
		return err
	}

	logger.Info("[VIDEO] Loading video file: " + p.config.FilePath)

	if err := p.player.Command([]string{"loadfile", p.config.FilePath}); err != nil {
		return err
	}

	// Start video playback (unpause video)
	if err := p.setPauseStatus(false); err != nil {
		logger.Error("[VIDEO] Failed to start video playback: " + err.Error())
	}

	// Start interval timer for updating video speed
	ticker := time.NewTicker(time.Second * time.Duration(p.config.UpdateIntervalSec))
	defer ticker.Stop()

	lastSpeed := 0.0

	logger.Info("[VIDEO] Entering MPV playback loop...")

	// Main loop for updating video speed
	for {

		select {
		case <-ctx.Done():
			logger.Info("[VIDEO] Context cancelled. Shutting down video player component")
			return nil
		case <-ticker.C:
			currentSpeed := speedController.GetSmoothedSpeed()

			logger.Info("[VIDEO] New sensor speed: " + strconv.FormatFloat(currentSpeed, 'f', 2, 64) + "[VIDEO] Last sensor speed: " + strconv.FormatFloat(lastSpeed, 'f', 2, 64))

			// Pause video if no speed is detected
			if currentSpeed == 0 {
				logger.Info("[VIDEO] No speed detected, so pausing video...")

				if err := p.setPauseStatus(true); err != nil {
					logger.Info("[VIDEO] Failed to pause video playback: " + err.Error())
				}

				continue
			}

			// Adjust playback speed if speed has changed beyond threshold value
			if math.Abs(currentSpeed-lastSpeed) > p.speedConfig.SpeedThreshold {
				newSpeed := (currentSpeed * p.config.SpeedMultiplier) / 10.0

				logger.Info("[VIDEO] Adjusting video speed to " + strconv.FormatFloat(newSpeed, 'f', 2, 64))

				if err := p.player.SetProperty("speed", mpv.FormatDouble, newSpeed); err != nil {
					logger.Error("[VIDEO] Failed to update video speed: " + err.Error())
				} else {
					logger.Info("[VIDEO] Video speed updated successfully")
				}

				// Display speed on mpv OSD (optional)
				if p.config.DisplaySpeed {

					if err := p.player.SetOptionString("osd-msg1", "Speed: "+fmt.Sprintf("%.2f", newSpeed)); err != nil {
						return err
					}

				}

				// Resume playback if paused
				if err := p.setPauseStatus(false); err != nil {
					logger.Error("[VIDEO] Failed to resume video playback: " + err.Error())
				}

			}

			lastSpeed = currentSpeed
		}
	}

}

// setPauseStatus sets the pause status of the video player to the specified value
func (p *PlaybackController) setPauseStatus(pause bool) error {

	// Check the current pause status
	currentPause, err := p.player.GetProperty("pause", mpv.FormatFlag)
	if err != nil {
		logger.Error("[VIDEO] Failed to get pause status: " + err.Error())
		return err
	}

	// If the pause status is already the desired state, do nothing
	if currentPause.(bool) == pause {
		return nil
	}

	// Set the pause status
	if err := p.player.SetProperty("pause", mpv.FormatFlag, pause); err != nil {
		result := func() string {
			if pause {
				return "paused"
			}
			return "resumed"
		}()
		logger.Error("[VIDEO] Failed to " + result + " video: " + err.Error())
		return err
	}

	// Log the success message... and punch out
	result := func() string {
		if pause {
			return "paused"
		}
		return "resumed"
	}()
	logger.Info("[VIDEO] Video " + result + " successfully")

	return nil

}
