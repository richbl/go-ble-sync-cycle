package video

import (
	"context"
	"log"
	"math"
	"time"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
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

	log.Println("/ Starting MPV video player...")

	if err := p.player.SetOption("window-scale", mpv.FormatDouble, 0.5); err != nil {
		return err
	}

	log.Println("/ Loading video file:", p.config.FilePath)

	if err := p.player.Command([]string{"loadfile", p.config.FilePath}); err != nil {
		return err
	}

	// Start video playback (unpause video)
	if err := p.setPauseStatus(false); err != nil {
		log.Println("/ Failed to start video playback:", err)
	}

	// Start interval timer for updating video speed
	ticker := time.NewTicker(time.Second * time.Duration(p.config.UpdateIntervalSec))
	defer ticker.Stop()

	lastSpeed := 0.0

	log.Println("/ Entering MPV playback loop...")

	// Main loop for updating video speed
	for {

		select {
		case <-ctx.Done():
			log.Println("/ Context cancelled. Shutting down video player component")
			return nil
		case <-ticker.C:
			currentSpeed := speedController.GetSmoothedSpeed()
			log.Printf("/ Current sensor speed: %.2f ... Last sensor speed: %.2f\n", currentSpeed, lastSpeed)

			// Pause video if no speed is detected
			if currentSpeed == 0 {
				log.Println("/ No speed detected, so pausing video...")

				if err := p.setPauseStatus(true); err != nil {
					log.Println("/ Failed to pause video playback:", err)
				}

				continue
			}

			// Adjust playback speed if speed has changed beyond threshold value
			if math.Abs(currentSpeed-lastSpeed) > p.speedConfig.SpeedThreshold {
				newSpeed := (currentSpeed * p.config.SpeedMultiplier) / 10.0

				log.Printf("/ Adjusting video speed to %.2f\n", newSpeed)

				if err := p.player.SetProperty("speed", mpv.FormatDouble, newSpeed); err != nil {
					log.Println("/ Failed to update video speed:", err)
				} else {
					log.Println("/ Video speed updated successfully")
				}

				// Resume playback if paused
				if err := p.setPauseStatus(false); err != nil {
					log.Println("/ Failed to resume video playback:", err)
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
		log.Println("/ Failed to get pause status:", err)
		return err
	}

	// If the pause status is already the desired state, do nothing
	if currentPause.(bool) == pause {
		return nil
	}

	// Set the pause status
	if err := p.player.SetProperty("pause", mpv.FormatFlag, pause); err != nil {
		log.Printf("/ Failed to %s video: %v", func() string {
			if pause {
				return "pause"
			}
			return "resume"
		}(), err)
		return err
	}

	// Log the success message... and punch out
	log.Printf("/ Video %s successfully", func() string {
		if pause {
			return "paused"
		}
		return "resumed"
	}())

	return nil

}
