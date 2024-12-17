package ble

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"tinygo.org/x/bluetooth"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

// Constants for speed calculations and BLE data parsing
const (
	minDataLength = 7
	wheelRevFlag  = uint8(0x01)
	kphConversion = 3.6
	mphConversion = 2.23694
)

// SpeedMeasurement represents the wheel revolution and time data from a BLE sensor
type SpeedMeasurement struct {
	wheelRevs uint32
	wheelTime uint16
}

// BLEController represents the BLE central controller component
type BLEController struct {
	bleConfig   config.BLEConfig
	speedConfig config.SpeedConfig
	bleAdapter  bluetooth.Adapter
}

// Package-level variables for tracking speed measurements
var (
	lastWheelRevs uint32
	lastWheelTime uint16
)

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*BLEController, error) {

	// Enable BLE adapter
	bleAdapter := bluetooth.DefaultAdapter
	if err := bleAdapter.Enable(); err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "created new BLE central controller")

	return &BLEController{
		bleConfig:   bleConfig,
		speedConfig: speedConfig,
		bleAdapter:  *bleAdapter,
	}, nil
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified UUID
func (m *BLEController) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	// Create context with timeout
	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(m.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	found := make(chan bluetooth.ScanResult, 1)
	errChan := make(chan error, 1)

	go func() {
		logger.Info(logger.BLE, "now scanning the ether for BLE peripheral UUID of "+m.bleConfig.SensorUUID+"...")

		if err := m.startScanning(found); err != nil {
			errChan <- err
		}
	}()

	// Wait for device discovery or timeout
	select {
	case result := <-found:
		logger.Debug(logger.BLE, "found BLE peripheral "+result.Address.String())
		return result, nil
	case err := <-errChan:
		return bluetooth.ScanResult{}, err
	case <-scanCtx.Done():
		if err := m.bleAdapter.StopScan(); err != nil {
			logger.Error(logger.BLE, "failed to stop scan: "+err.Error())
		}
		return bluetooth.ScanResult{}, errors.New("scanning time limit reached")
	}
}

// startScanning starts the BLE scan and sends the result to the found channel when the target device is discovered
func (m *BLEController) startScanning(found chan<- bluetooth.ScanResult) error {

	// Start BLE scan
	err := m.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		// Check if the target peripheral was found
		if result.Address.String() == m.bleConfig.SensorUUID {
			// Stop scanning
			if err := m.bleAdapter.StopScan(); err != nil {
				logger.Error(fmt.Sprintf(string(logger.BLE)+"failed to stop scan: %v", err))
			}

			// Found the target peripheral
			found <- result
		}
	})
	if err != nil {
		logger.Error(logger.BLE, "scan error: "+err.Error())
	}

	return nil
}

// GetBLECharacteristic scans for the BLE peripheral and returns CSC services/characteristics
func (m *BLEController) GetBLECharacteristic(ctx context.Context, speedController *speed.SpeedController) (*bluetooth.DeviceCharacteristic, error) {

	// Scan for BLE peripheral
	result, err := m.ScanForBLEPeripheral(ctx)
	if err != nil {
		return nil, err
	}

	logger.Debug(logger.BLE, "connecting to BLE peripheral device "+result.Address.String())

	// Connect to BLE peripheral device
	var device bluetooth.Device
	if device, err = m.bleAdapter.Connect(result.Address, bluetooth.ConnectionParams{}); err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "BLE peripheral device connected")
	logger.Debug(logger.BLE, "discovering CSC services "+bluetooth.New16BitUUID(0x1816).String())

	// Find CSC service and characteristic
	svc, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0x1816)})
	if err != nil {
		logger.Error(logger.BLE, "CSC services discovery failed: "+err.Error())
		return nil, err
	}

	logger.Debug(logger.BLE, "found CSC service "+svc[0].UUID().String())
	logger.Debug(logger.BLE, "discovering CSC characteristics "+bluetooth.New16BitUUID(0x2A5B).String())

	char, err := svc[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0x2A5B)})
	if err != nil {
		logger.Warn(logger.BLE, "CSC characteristics discovery failed: "+err.Error())
		return nil, err
	}

	logger.Debug(logger.BLE, "found CSC characteristic "+char[0].UUID().String())
	return &char[0], nil
}

// GetBLEUpdates enables BLE peripheral monitoring to report real-time sensor data
func (m *BLEController) GetBLEUpdates(ctx context.Context, speedController *speed.SpeedController, char *bluetooth.DeviceCharacteristic) error {

	logger.Debug(logger.BLE, "starting real-time monitoring of BLE sensor notifications...")

	// Subscribe to live BLE sensor notifications
	if err := char.EnableNotifications(func(buf []byte) {
		speed := m.ProcessBLESpeed(buf)
		speedController.UpdateSpeed(speed)
	}); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// ProcessBLESpeed processes the raw speed data from the BLE peripheral
func (m *BLEController) ProcessBLESpeed(data []byte) float64 {

	// Parse speed data
	newSpeedData, err := m.parseSpeedData(data)
	if err != nil {
		logger.Error(logger.SPEED, "invalid BLE data: "+err.Error())
		return 0.0
	}

	// Calculate speed from parsed data
	speed := m.calculateSpeed(newSpeedData)
	logger.Info(logger.SPEED, logger.Blue+"BLE sensor speed: "+strconv.FormatFloat(speed, 'f', 2, 64)+" "+m.speedConfig.SpeedUnits)

	return speed
}

// calculateSpeed calculates the current speed based on the sensor data
func (m *BLEController) calculateSpeed(sm SpeedMeasurement) float64 {

	// First time through the loop set the last wheel revs and time
	if lastWheelTime == 0 {
		lastWheelRevs = sm.wheelRevs
		lastWheelTime = sm.wheelTime
		return 0.0
	}

	// Calculate delta between time intervals
	timeDiff := sm.wheelTime - lastWheelTime
	if timeDiff == 0 {
		return 0.0
	}

	// Calculate delta between wheel revs
	revDiff := int32(sm.wheelRevs - lastWheelRevs)

	// Determine speed unit conversion multiplier
	speedConversion := kphConversion
	if m.speedConfig.SpeedUnits == config.SpeedUnitsMPH {
		speedConversion = mphConversion
	}

	// Calculate new speed
	speed := float64(revDiff) * float64(m.speedConfig.WheelCircumferenceMM) * speedConversion / float64(timeDiff)
	lastWheelRevs = sm.wheelRevs
	lastWheelTime = sm.wheelTime

	return speed
}

// parseSpeedData parses the raw speed data from the BLE peripheral
func (m *BLEController) parseSpeedData(data []byte) (SpeedMeasurement, error) {

	// Check for data
	if len(data) < 1 {
		return SpeedMeasurement{}, errors.New("empty data")
	}

	// Validate data
	if data[0]&wheelRevFlag == 0 || len(data) < minDataLength {
		return SpeedMeasurement{}, errors.New("invalid data format or length")
	}

	// Return new speed data
	return SpeedMeasurement{
		wheelRevs: binary.LittleEndian.Uint32(data[1:]),
		wheelTime: binary.LittleEndian.Uint16(data[5:]),
	}, nil
}
