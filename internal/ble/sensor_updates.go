package ble

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

const (
	minDataLength = 7
	wheelRevFlag  = uint8(0x01)
	kphConversion = 3.6     // Conversion factor for kilometers per hour
	mphConversion = 2.23694 // Conversion factor for miles per hour
)

// SpeedMeasurement represents the values needed to calculate the speed
type SpeedMeasurement struct {
	wheelRevs uint32
	wheelTime uint16
}

// GetBLEUpdates starts the real-time monitoring of BLE sensor notifications
func (m *BLEController) GetBLEUpdates(ctx context.Context, speedController *speed.SpeedController) error {

	logger.Info(logger.BLE, "starting real-time monitoring of BLE sensor notifications...")
	errChan := make(chan error, 1)

	if err := m.bleDetails.bleCharacteristic.EnableNotifications(func(buf []byte) {
		speed := m.ProcessBLESpeed(buf)
		speedController.UpdateSpeed(speed)
	}); err != nil {
		return err
	}

	// Disable notifications after the context is canceled
	defer func() {

		if err := m.bleDetails.bleCharacteristic.EnableNotifications(nil); err != nil {
			logger.Error(logger.BLE, "failed to disable notifications:", err.Error())
		}

	}()

	go func() {
		<-ctx.Done()
		fmt.Print("\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE peripheral reporting...")
		errChan <- nil
	}()

	return <-errChan
}

// ProcessBLESpeed processes raw BLE speed data into human-readable speed values
func (m *BLEController) ProcessBLESpeed(data []byte) float64 {

	newSpeedData, err := m.parseSpeedData(data)
	if err != nil {
		logger.Error(logger.SPEED, "invalid BLE data:", err.Error())
		return 0.0
	}

	//
	speed := m.calculateSpeed(newSpeedData)
	logger.Debug(logger.SPEED, logger.Blue+"BLE sensor speed:", strconv.FormatFloat(speed, 'f', 2, 64), m.speedConfig.SpeedUnits)

	return speed
}

// calculateSpeed calculates the speed from the raw BLE data
func (m *BLEController) calculateSpeed(sm SpeedMeasurement) float64 {

	// Initialize the last wheel revs and time
	if m.lastWheelTime == 0 {
		m.lastWheelRevs = sm.wheelRevs
		m.lastWheelTime = sm.wheelTime
		return 0.0
	}

	// Get the time difference between the current and last wheel revs
	timeDiff := sm.wheelTime - m.lastWheelTime
	if timeDiff == 0 {
		return 0.0
	}

	// Get the rev difference between the current and last wheel revs
	revDiff := int32(sm.wheelRevs - m.lastWheelRevs)
	speedConversion := kphConversion

	if m.speedConfig.SpeedUnits == config.SpeedUnitsMPH {
		speedConversion = mphConversion
	}

	// Calculate the speed in km/h or mph
	speed := float64(revDiff) * float64(m.speedConfig.WheelCircumferenceMM) * speedConversion / float64(timeDiff)
	m.lastWheelRevs = sm.wheelRevs
	m.lastWheelTime = sm.wheelTime

	return speed
}

// parseSpeedData parses the raw BLE speed data
func (m *BLEController) parseSpeedData(data []byte) (SpeedMeasurement, error) {

	if len(data) < 1 {
		return SpeedMeasurement{}, fmt.Errorf("no speed data reported")
	}

	if data[0]&wheelRevFlag == 0 || len(data) < minDataLength {
		return SpeedMeasurement{}, fmt.Errorf("invalid data format or length")
	}

	return SpeedMeasurement{
		wheelRevs: binary.LittleEndian.Uint32(data[1:]),
		wheelTime: binary.LittleEndian.Uint16(data[5:]),
	}, nil
}
