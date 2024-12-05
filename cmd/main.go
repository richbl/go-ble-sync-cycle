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
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
	video "github.com/richbl/go-ble-sync-cycle/internal/video-player"

	"tinygo.org/x/bluetooth"
)

func main() {

	log.Println("Starting BLE Sync Cycle 0.5.0")

	// Load configuration file (TOML)
	cfg, err := config.LoadFile("internal/configuration/config.toml")
	if err != nil {
		log.Fatal("FATAL - Failed to load TOML configuration: " + err.Error())
	}

	// Initialize logger
	logger.Initialize(cfg.App.LogLevel)

	// Create contexts to manage goroutines and system interrupts
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	ctx, stop := signal.NotifyContext(rootCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create component controllers
	speedController, videoPlayer, bleController, err := createControllers(*cfg)
	if err != nil {
		logger.Fatal("[APP] Failed to create controllers: " + err.Error())
	}

	// Scan for BLE peripheral and return CSC speed characteristic
	bleSpeedCharacter, err := scanForBLESpeedCharacteristic(ctx, speedController, bleController)
	if err != nil {
		logger.Error("[BLE] BLE peripheral scan failed: " + err.Error())
		return
	}

	// Start components (running concurrently)
	var wg sync.WaitGroup

	// Start BLE peripheral speed monitoring
	if err := monitorBLESpeed(ctx, &wg, bleController, speedController, bleSpeedCharacter, rootCancel); err != nil {
		logger.Error("[BLE] Failed to start BLE speed monitoring: ", err.Error())
		return
	}

	// Start video playback
	if err := playVideo(ctx, &wg, videoPlayer, speedController, rootCancel); err != nil {
		logger.Error("[VIDEO] Failed to start video playback: ", err.Error())
		return
	}

	// Set up interrupt handling, allowing for user interrupts and graceful component shutdown
	if err := interruptHandler(ctx, rootCancel); err != nil {
		logger.Error("[APP] Failed to set up interrupt handling: ", err.Error())
		return
	}

	// Wait for all goroutines to complete
	wg.Wait()
	logger.Info("[APP] Application shutdown complete. Goodbye!")

}

// createControllers creates the speed controller, video player, and BLE controller
func createControllers(cfg config.Config) (*speed.SpeedController, *video.PlaybackController, *ble.BLEController, error) {

	// Create speed controller component
	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)

	// Create video player component
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create BLE controller component
	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return nil, nil, nil, err
	}

	return speedController, videoPlayer, bleController, nil

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
				logger.Error("[BLE] BLE speed characteristic scan cancelled: ", ctx.Err())
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
			logger.Error("[BLE] BLE speed characteristic monitoring error: ", err.Error())
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
			logger.Error("[VIDEO] Video playback error: ", err.Error())
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
			logger.Info("[APP] Shutdown signal received")
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
