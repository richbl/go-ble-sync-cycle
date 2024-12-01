package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	ble "github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

func main() {

	// Disable logging
	// log.SetOutput(io.Discard)

	log.Println("- Starting BLE Sync Cycle 0.5.0")

	// Load configuration file (TOML)
	cfg, err := config.LoadFile("internal/configuration/config.toml")
	if err != nil {
		log.Fatalln("- Failed to load configuration:", err)
	}

	// Verify video file exists for playback
	if _, err = os.Stat(cfg.Video.FilePath); os.IsNotExist(err) {
		log.Fatalln("- Video file does not exist:", err)
	}

	// Create contexts to manage goroutines and system interrupts
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	ctx, stop := signal.NotifyContext(rootCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create speed controller component
	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)

	// Create video player component
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		log.Fatalln("/ Failed to create video player component:", err)
	}

	// Create BLE controller component
	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		log.Fatalln("\\ Failed to create BLE controller component:", err)
	}

	// Scan for BLE peripheral and return CSC speed characteristic
	bleSpeedCharacter, err := scanForBLESpeedCharacteristic(ctx, speedController, bleController)
	if err != nil {
		log.Printf("\\ BLE peripheral scan failed: %v", err)
		return
	}

	// Start components (running concurrently)
	var wg sync.WaitGroup

	// Start BLE peripheral speed monitoring
	if err := monitorBLESpeed(ctx, &wg, bleController, speedController, bleSpeedCharacter, rootCancel); err != nil {
		log.Printf("\\ Failed to start BLE speed monitoring: %v", err)
		return
	}

	// Start video playback
	if err := playVideo(ctx, &wg, videoPlayer, speedController, rootCancel); err != nil {
		log.Printf("/ Failed to start video playback: %v", err)
		return
	}

	// Set up interrupt handling, allowing for user interrupts and graceful component shutdown
	if err := interruptHandler(ctx, rootCancel); err != nil {
		log.Printf("- Failed to set up interrupt handling: %v", err)
		return
	}

	// Wait for all goroutines to complete
	wg.Wait()
	log.Println("- Application shutdown complete. Goodbye!")

}

// scanForBLESpeedCharacteristic scans for the BLE peripheral and returns the CSC speed characteristic
func scanForBLESpeedCharacteristic(ctx context.Context, speedController *speed.SpeedController, bleController *ble.BLEController) (*bluetooth.DeviceCharacteristic, error) {

	var bleSpeedCharacter *bluetooth.DeviceCharacteristic
	errChan := make(chan error, 1)

	go func() {

		var err error
		bleSpeedCharacter, err = bleController.GetBLECharacteristic(ctx, speedController)
		errChan <- err

	}()

	// Wait for the scanning process to complete or context cancellation
	select {
	case err := <-errChan:

		if err != nil {

			if ctx.Err() != nil {
				log.Printf("\\ BLE speed characteristic scan cancelled: %v", ctx.Err())
				return nil, ctx.Err()
			}
			return nil, err

		}

		return bleSpeedCharacter, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	}
}

// monitorBLESpeed starts BLE speed characteristic monitoring used for reporting real-time sensor data
func monitorBLESpeed(ctx context.Context, wg *sync.WaitGroup, bleController *ble.BLEController, speedController *speed.SpeedController, bleSpeedCharacter *bluetooth.DeviceCharacteristic, cancel context.CancelFunc) error {

	errChan := make(chan error, 1)
	wg.Add(1)

	go func() {

		defer wg.Done()

		if err := bleController.GetBLEUpdates(ctx, speedController, bleSpeedCharacter); err != nil {
			log.Printf("\\ BLE speed characteristic monitoring error: %v", err)
			errChan <- err
			cancel()
			return
		}

		errChan <- nil

	}()

	// Wait for BLE monitoring to complete or context cancellation
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	default:
		return nil
	}

}

// playVideo starts the video player and monitors speed changes from the BLE sensor
func playVideo(ctx context.Context, wg *sync.WaitGroup, videoPlayer *video.PlaybackController, speedController *speed.SpeedController, cancel context.CancelFunc) error {

	errChan := make(chan error, 1)
	wg.Add(1)

	go func() {

		defer wg.Done()

		if err := videoPlayer.Start(ctx, speedController); err != nil {
			log.Printf("/ Video playback error: %v", err)
			errChan <- err
			cancel()
			return
		}

		errChan <- nil

	}()

	// Wait for video playback to complete or context cancellation
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	default:
		return nil
	}

}

// interruptHandler waits for shutdown signals and cancels the context when received
func interruptHandler(ctx context.Context, cancel context.CancelFunc) error {

	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)

	go func() {

		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigChan:
			log.Println("- Shutdown signal received")
			cancel()
			errChan <- nil
		case <-ctx.Done():
			errChan <- nil
		}

	}()

	// Wait for video playback to complete or context cancellation
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	default:
		return nil
	}

}
