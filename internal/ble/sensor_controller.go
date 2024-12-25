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

// Constants for BLE data parsing and speed calculations
const (
	minDataLength = 7
	wheelRevFlag  = uint8(0x01)
	kphConversion = 3.6     // Conversion factor for kilometers per hour
	mphConversion = 2.23694 // Conversion factor for miles per hour
)

// SpeedMeasurement represents the wheel revolution and time data from a BLE sensor
type SpeedMeasurement struct {
	wheelRevs uint32
	wheelTime uint16
}

// BLEController manages BLE communication with cycling speed sensors
type BLEController struct {
	bleConfig     config.BLEConfig
	speedConfig   config.SpeedConfig
	bleAdapter    bluetooth.Adapter
	lastWheelRevs uint32
	lastWheelTime uint16
}

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*BLEController, error) {

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

// GetBLECharacteristic scans for the BLE peripheral and returns CSC services/characteristics
func (m *BLEController) GetBLECharacteristic(ctx context.Context, speedController *speed.SpeedController) (*bluetooth.DeviceCharacteristic, error) {

	// Scan for BLE peripheral device
	result, err := m.ScanForBLEPeripheral(ctx)
	if err != nil {
		return nil, err
	}

	logger.Debug(logger.BLE, "connecting to BLE peripheral device", result.Address.String())

	device, err := m.bleAdapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "BLE peripheral device connected")
	logger.Debug(logger.BLE, "discovering CSC services", bluetooth.New16BitUUID(0x1816).String())

	// Discover CSC services
	svc, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0x1816)})
	if err != nil {
		logger.Error(logger.BLE, "CSC services discovery failed:", err.Error())
		return nil, err
	}

	logger.Debug(logger.BLE, "found CSC service", svc[0].UUID().String())
	logger.Debug(logger.BLE, "discovering CSC characteristics", bluetooth.New16BitUUID(0x2A5B).String())

	// Discover CSC characteristics
	char, err := svc[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0x2A5B)})
	if err != nil {
		logger.Warn(logger.BLE, "CSC characteristics discovery failed:", err.Error())
		return nil, err
	}

	logger.Debug(logger.BLE, "found CSC characteristic", char[0].UUID().String())
	return &char[0], nil
}

// GetBLEUpdates enables real-time monitoring of BLE peripheral sensor data, handling
// notification setup/teardown, and updates the speed controller with new readings
func (m *BLEController) GetBLEUpdates(ctx context.Context, speedController *speed.SpeedController, char *bluetooth.DeviceCharacteristic) error {

	logger.Debug(logger.BLE, "starting real-time monitoring of BLE sensor notifications...")
	errChan := make(chan error, 1)

	if err := char.EnableNotifications(func(buf []byte) {
		speed := m.ProcessBLESpeed(buf)
		speedController.UpdateSpeed(speed)
	}); err != nil {
		return err
	}

	// Need to disable BLE notifications when done
	defer func() {
		if err := char.EnableNotifications(nil); err != nil {
			logger.Error(logger.BLE, "failed to disable notifications:", err.Error())
		}
	}()

	go func() {
		<-ctx.Done()
		fmt.Print("\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE component reporting...")
		errChan <- nil
	}()

	return <-errChan
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified UUID
func (m *BLEController) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(m.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	found := make(chan bluetooth.ScanResult, 1)
	errChan := make(chan error, 1)

	go func() {
		logger.Info(logger.BLE, "now scanning the ether for BLE peripheral UUID of", m.bleConfig.SensorUUID+"...")
		if err := m.startScanning(found); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-found:
		logger.Debug(logger.BLE, "found BLE peripheral", result.Address.String())
		return result, nil
	case err := <-errChan:
		return bluetooth.ScanResult{}, err
	case <-scanCtx.Done():
		if err := m.bleAdapter.StopScan(); err != nil {
			logger.Error(logger.BLE, "failed to stop scan:", err.Error())
		}

		return bluetooth.ScanResult{}, errors.New("scanning time limit reached")
	}
}

// ProcessBLESpeed processes raw speed data from the BLE peripheral and returns the calculated speed
func (m *BLEController) ProcessBLESpeed(data []byte) float64 {

	newSpeedData, err := m.parseSpeedData(data)
	if err != nil {
		logger.Error(logger.SPEED, "invalid BLE data:", err.Error())
		return 0.0
	}

	speed := m.calculateSpeed(newSpeedData)
	logger.Debug(logger.SPEED, logger.Blue+"BLE sensor speed:", strconv.FormatFloat(speed, 'f', 2, 64), m.speedConfig.SpeedUnits)

	return speed
}

// startScanning starts the BLE scan and sends results to the found channel
func (m *BLEController) startScanning(found chan<- bluetooth.ScanResult) error {

	err := m.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

		if result.Address.String() == m.bleConfig.SensorUUID {

			if err := m.bleAdapter.StopScan(); err != nil {
				logger.Error(logger.BLE, "failed to stop scan:", err.Error())
			}

			found <- result
		}

	})

	if err != nil {
		logger.Error(logger.BLE, "scan error:", err.Error())
	}

	return nil
}

// calculateSpeed calculates the current speed based on wheel revolution data... interestingly,
// a BLE speed sensor has no concept of rate: just wheel revolutions and timestamps
func (m *BLEController) calculateSpeed(sm SpeedMeasurement) float64 {

	// Initialize last wheel data if not set
	if m.lastWheelTime == 0 {
		m.lastWheelRevs = sm.wheelRevs
		m.lastWheelTime = sm.wheelTime
		return 0.0
	}

	timeDiff := sm.wheelTime - m.lastWheelTime
	if timeDiff == 0 {
		return 0.0
	}

	revDiff := int32(sm.wheelRevs - m.lastWheelRevs)
	speedConversion := kphConversion
	if m.speedConfig.SpeedUnits == config.SpeedUnitsMPH {
		speedConversion = mphConversion
	}

	speed := float64(revDiff) * float64(m.speedConfig.WheelCircumferenceMM) * speedConversion / float64(timeDiff)
	m.lastWheelRevs = sm.wheelRevs
	m.lastWheelTime = sm.wheelTime

	return speed
}

// parseSpeedData parses raw byte data from the BLE peripheral into a SpeedMeasurement
func (m *BLEController) parseSpeedData(data []byte) (SpeedMeasurement, error) {

	if len(data) < 1 {
		return SpeedMeasurement{}, errors.New("empty data")
	}

	if data[0]&wheelRevFlag == 0 || len(data) < minDataLength {
		return SpeedMeasurement{}, errors.New("invalid data format or length")
	}

	return SpeedMeasurement{
		wheelRevs: binary.LittleEndian.Uint32(data[1:]),
		wheelTime: binary.LittleEndian.Uint16(data[5:]),
	}, nil
}
