package ble

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

const (
	minDataLength = 7           // Data length as defined in BLE CSC specification
	wheelRevFlag  = uint8(0x01) // Wheel revolutions flag as defined in BLE CSC specification
	mphConversion = 0.621371    // Conversion factor for miles per hour
)

// speedData represents the values needed to calculate the speed
type speedData struct {
	wheelTime     uint16
	lastWheelTime uint16
	wheelRevs     uint32
	lastWheelRevs uint32
	distance      float64

	// Pre-calculated speed constants
	wheelCircumferenceM   float64 // wheelCircumferenceMM / 1000
	timeConversionFactor  float64 // 1/1024 seconds (BLE CSC specification time interval)
	speedConversionFactor float64 // 3.6 * speedUnitMultiplier
}

// unitConversion maps speed units to their respective conversion factors
var unitConversion = map[string]float64{
	config.SpeedUnitsKMH: 1.0,
	config.SpeedUnitsMPH: mphConversion,
}

// initSpeedData initializes the speedData struct with pre-calculated constants
func initSpeedData(wheelCircumferenceMM int, speedUnitMultiplier float64) *speedData {
	return &speedData{
		wheelCircumferenceM:   float64(wheelCircumferenceMM) / 1000,
		timeConversionFactor:  1.0 / 1024,
		speedConversionFactor: 3.6 * speedUnitMultiplier,
	}
}

// GetBLEUpdates starts the real-time monitoring of BLE sensor notifications
func (m *Controller) GetBLEUpdates(ctx context.Context, speedController *speed.Controller) error {

	logger.Info(logger.BLE, "starting the monitoring for BLE sensor notifications...")

	errChan := make(chan error, 1)

	// Precalculate speed data values
	speedUnitMultiplier := unitConversion[m.speedConfig.SpeedUnits]
	sd := initSpeedData(m.speedConfig.WheelCircumferenceMM, speedUnitMultiplier)

	notificationHandler := func(buf []byte) {
		speed, err := sd.processBLESpeed(m.speedConfig.SpeedUnits, buf)
		if err != nil {
			logger.Warn(logger.SPEED, fmt.Sprintf("error processing BLE speed data: %v", err))
			return
		}
		speedController.UpdateSpeed(speed)
	}

	// Enable real-time notifications from BLE sensor
	if err := m.blePeripheralDetails.bleCharacteristic.EnableNotifications(notificationHandler); err != nil {
		return fmt.Errorf(errFormat, ErrNotificationEnable, err)
	}

	// Manage context cancellation
	go func() {
		<-ctx.Done()
		fmt.Print("\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BLE, "interrupt detected, stopping the monitoring for BLE sensor notifications...")

		// Disable real-time notifications from BLE sensor
		if err := m.blePeripheralDetails.bleCharacteristic.EnableNotifications(nil); err != nil {
			logger.Error(logger.BLE, fmt.Sprintf("failed to disable BLE notifications: %v", err))
		}

		errChan <- nil
		close(errChan)
	}()

	return <-errChan
}

// processBLESpeed processes raw BLE speed data into human-readable speed values
func (sd *speedData) processBLESpeed(speedUnits string, speedData []byte) (float64, error) {

	if err := sd.parseSpeedData(speedData); err != nil {
		return 0.0, err
	}

	speed := sd.calculateSpeed()
	logger.Debug(logger.SPEED, fmt.Sprintf("%sBLE sensor speed: %.2f %s", logger.Blue, speed, speedUnits))

	return speed, nil
}

// calculateSpeed calculates the speed from the raw BLE data
func (sd *speedData) calculateSpeed() float64 {

	// Initialize last wheel revs and time if they are zero
	if sd.lastWheelTime == 0 {
		return sd.initializeWheelData()
	}

	// Get the rev and time differences (in 1/1024 seconds) between the current and last wheel revs
	revDiff := sd.wheelRevs - sd.lastWheelRevs
	timeDiff := sd.wheelTime - sd.lastWheelTime

	// Early exit if no data has changed
	if timeDiff == 0 || revDiff == 0 {
		return 0.0
	}

	// Calculate the distance (in meters)
	distance := float64(revDiff) * sd.wheelCircumferenceM

	// Update the total distance cycled
	sd.distance += distance

	// Calculate the speed in km/h or mph
	speed := (distance / (float64(timeDiff) * sd.timeConversionFactor)) * sd.speedConversionFactor

	// Round the speed to two decimal places
	speed = math.Round(speed*100) / 100

	// Update the last values for next calculation
	sd.lastWheelRevs = sd.wheelRevs
	sd.lastWheelTime = sd.wheelTime

	return speed
}

// initializeWheelData initializes the speed data
func (sd *speedData) initializeWheelData() float64 {

	sd.lastWheelRevs = sd.wheelRevs
	sd.lastWheelTime = sd.wheelTime

	return 0.0
}

// parseSpeedData parses the raw BLE speed data
func (sd *speedData) parseSpeedData(speedData []byte) error {

	if len(speedData) < 1 {
		return ErrNoSpeedData
	}

	if speedData[0]&wheelRevFlag == 0 || len(speedData) < minDataLength {
		return ErrInvalidSpeedData
	}

	sd.wheelRevs = binary.LittleEndian.Uint32(speedData[1:5])
	sd.wheelTime = binary.LittleEndian.Uint16(speedData[5:7])

	return nil
}
