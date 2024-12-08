package main

import (
	"context"
	"errors"
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

// appControllers holds the main application controllers
type appControllers struct {
	speedController *speed.SpeedController
	videoPlayer     *video.PlaybackController
	bleController   *ble.BLEController
}

func main() {

	log.Println("Starting BLE Sync Cycle 0.5.0")

	// Load configuration
	cfg, err := config.LoadFile("internal/configuration/config.toml")
	if err != nil {
		log.Fatal("FATAL - Failed to load TOML configuration: " + err.Error())
	}

	// Initialize logger
	logger.Initialize(cfg.App.LogLevel)

	// Create contexts for managing goroutines and cancellations
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Create component controllers
	controllers, err := setupAppControllers(*cfg)
	if err != nil {
		logger.Fatal("[APP] Failed to create controllers: " + err.Error())
	}

	// Run the application
	if err := startAppControllers(rootCtx, controllers); err != nil {
		logger.Error(err.Error())
	}

	// Shutdown the application... buh bye!
	logger.Info("[APP] Application shutdown complete. Goodbye!")

}

// startAppControllers is responsible for starting and managing the component controllers
func startAppControllers(ctx context.Context, controllers appControllers) error {

	// Create shutdown signal
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Scan for BLE peripheral of interest
	bleSpeedCharacter, err := scanForBLESpeedCharacteristic(ctx, controllers)
	if err != nil {
		return errors.New("[BLE] BLE peripheral scan failed: " + err.Error())
	}

	// Start component controllers concurrently
	var wg sync.WaitGroup
	errs := make(chan error, 2)

	// Start BLE sensor speed monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := monitorBLESpeed(ctx, controllers, bleSpeedCharacter); err != nil {
			errs <- errors.New("[BLE] Speed monitoring failed: " + err.Error())
		}
	}()

	// Start video playback
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := playVideo(ctx, controllers); err != nil {
			errs <- errors.New("[VIDEO] Playback failed: " + err.Error())
		}
	}()

	// Wait for shutdown signals from go routines
	go func() {
		<-ctx.Done()
		logger.Info("[APP] Shutdown signal received")
	}()

	// Wait for goroutines to finish or an error to occur
	go func() {
		wg.Wait()
		close(errs)
	}()

	// Check for errors from components
	if err := <-errs; err != nil {
		return err
	}

	return nil
}

// setupAppControllers creates and initializes the application controllers
func setupAppControllers(cfg config.Config) (appControllers, error) {

	// Create speed controller
	speedController := speed.NewSpeedController(cfg.Speed.SmoothingWindow)

	// Create video player
	videoPlayer, err := video.NewPlaybackController(cfg.Video, cfg.Speed)
	if err != nil {
		return appControllers{}, errors.New("[VIDEO] Failed to create video player: " + err.Error())
	}

	// Create BLE controller
	bleController, err := ble.NewBLEController(cfg.BLE, cfg.Speed)
	if err != nil {
		return appControllers{}, errors.New("[BLE] Failed to create BLE controller: " + err.Error())
	}

	return appControllers{
		speedController: speedController,
		videoPlayer:     videoPlayer,
		bleController:   bleController,
	}, nil

}

// scanForBLESpeedCharacteristic scans for the BLE CSC speed characteristic
func scanForBLESpeedCharacteristic(ctx context.Context, controllers appControllers) (*bluetooth.DeviceCharacteristic, error) {

	// create a channel to receive the characteristic
	results := make(chan *bluetooth.DeviceCharacteristic, 1)
	errChan := make(chan error, 1)

	// Scan for the BLE CSC speed characteristic
	go func() {
		characteristic, err := controllers.bleController.GetBLECharacteristic(ctx, controllers.speedController)

		if err != nil {
			errChan <- err
			return
		}

		// Return the characteristic
		results <- characteristic

	}()

	// Wait for the characteristic or an error
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case characteristic := <-results:
		return characteristic, nil
	}

}

// monitorBLESpeed monitors the BLE speed characteristic
func monitorBLESpeed(ctx context.Context, controllers appControllers, bleSpeedCharacter *bluetooth.DeviceCharacteristic) error {

	return controllers.bleController.GetBLEUpdates(ctx, controllers.speedController, bleSpeedCharacter)

}

// playVideo starts the video player
func playVideo(ctx context.Context, controllers appControllers) error {

	return controllers.videoPlayer.Start(ctx, controllers.speedController)

}
