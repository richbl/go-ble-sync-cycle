package ble

import (
	"context"
	"encoding/binary"
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

// BLEDetails holds BLE peripheral details
type BLEDetails struct {
	bleConfig         config.BLEConfig
	bleAdapter        bluetooth.Adapter
	bleCharacteristic *bluetooth.DeviceCharacteristic
}

// BLEController holds the BLE controller component and sensor data
type BLEController struct {
	bleDetails    BLEDetails
	speedConfig   config.SpeedConfig
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
		bleDetails: BLEDetails{
			bleConfig:  bleConfig,
			bleAdapter: *bleAdapter,
		},
		speedConfig: speedConfig,
	}, nil
}

// performBLEAction performs the provided BLE setup action
func (m *BLEController) performBLEAction(ctx context.Context, action func(found chan<- interface{}, errChan chan<- error), logMessage string, stopAction func() error) (interface{}, error) {

	// Create a context with a timeout for the scan
	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(m.bleDetails.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	found := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	// Run the action in a goroutine and handle the results
	go func() {
		logger.Debug(logger.BLE, logMessage)
		action(found, errChan)
	}()

	select {
	case result := <-found:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-scanCtx.Done():

		if stopAction != nil {

			if err := stopAction(); err != nil {
				fmt.Print("\r") // Clear the ^C character from the terminal line
				logger.Error(logger.BLE, "failed to stop action:", err.Error())
			}

		}

		if scanCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("scanning time limit reached")
		}

		fmt.Print("\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BLE, "user-generated interrupt, stopping BLE device setup...")
		return nil, scanCtx.Err()
	}

}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified UUID
func (m *BLEController) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	// Pass anonymous function into performBLEAction to scan for BLE peripheral
	result, err := m.performBLEAction(ctx, func(found chan<- interface{}, errChan chan<- error) {

		foundChan := make(chan bluetooth.ScanResult, 1)

		// Start scanning for BLE peripherals
		if err := m.startScanning(foundChan); err != nil {
			errChan <- err
			return
		}

		found <- <-foundChan
	}, fmt.Sprintf("scanning for BLE peripheral UUID %s", m.bleDetails.bleConfig.SensorUUID), m.bleDetails.bleAdapter.StopScan)
	if err != nil {
		return bluetooth.ScanResult{}, err
	}

	typedResult := result.(bluetooth.ScanResult)
	logger.Info(logger.BLE, "found BLE peripheral", typedResult.Address.String())
	return typedResult, nil
}

// ConnectToBLEPeripheral connects to the specified BLE peripheral
func (m *BLEController) ConnectToBLEPeripheral(ctx context.Context, device bluetooth.ScanResult) (bluetooth.Device, error) {

	// Pass anonymous function into performBLEAction to connect to BLE peripheral
	result, err := m.performBLEAction(ctx, func(found chan<- interface{}, errChan chan<- error) {

		// Connect to the BLE peripheral
		dev, err := m.bleDetails.bleAdapter.Connect(device.Address, bluetooth.ConnectionParams{})

		if err != nil {
			errChan <- err
			return
		}

		found <- dev
	}, fmt.Sprintf("connecting to BLE peripheral %s", device.Address.String()), nil)
	if err != nil {
		return bluetooth.Device{}, err
	}

	typedResult := result.(bluetooth.Device)
	logger.Info(logger.BLE, "BLE peripheral device connected")
	return typedResult, nil
}

// GetBLEServiceCharacteristic retrieves CSC services from the BLE peripheral
func (m *BLEController) GetBLEServices(ctx context.Context, device bluetooth.Device) ([]bluetooth.DeviceService, error) {

	// Pass anonymous function into performBLEAction to discover CSC services
	result, err := m.performBLEAction(ctx, func(found chan<- interface{}, errChan chan<- error) {

		// Discover CSC services
		services, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0x1816)})

		if err != nil {
			errChan <- err
			return
		}

		found <- services
	}, "discovering CSC service "+bluetooth.New16BitUUID(0x1816).String(), nil)
	if err != nil {
		return nil, err
	}

	typedResult := result.([]bluetooth.DeviceService)
	logger.Info(logger.BLE, "found CSC service", typedResult[0].UUID().String())
	return typedResult, nil
}

// GetBLECharacteristics retrieves CSC characteristics from the BLE peripheral
func (m *BLEController) GetBLECharacteristics(ctx context.Context, services []bluetooth.DeviceService) error {

	// Pass anonymous function into performBLEAction to discover CSC characteristics
	_, err := m.performBLEAction(ctx, func(found chan<- interface{}, errChan chan<- error) {

		// Discover CSC characteristics
		characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0x2A5B)})

		if err != nil {
			errChan <- err
			return
		}

		m.bleDetails.bleCharacteristic = &characteristics[0]
		found <- characteristics
	}, "discovering CSC characteristic "+bluetooth.New16BitUUID(0x2A5B).String(), nil)
	if err != nil {
		logger.Error(logger.BLE, "CSC characteristics discovery failed:", err.Error())
		return err
	}

	logger.Info(logger.BLE, "found CSC characteristic", m.bleDetails.bleCharacteristic.UUID().String())
	return nil
}

// GetBLEUpdates enables real-time monitoring of BLE peripheral sensor data, handling notification
// setup/teardown, and updates the speed controller with new readings
func (m *BLEController) GetBLEUpdates(ctx context.Context, speedController *speed.SpeedController) error {

	logger.Info(logger.BLE, "starting real-time monitoring of BLE sensor notifications...")
	errChan := make(chan error, 1)

	if err := m.bleDetails.bleCharacteristic.EnableNotifications(func(buf []byte) {
		speed := m.ProcessBLESpeed(buf)
		speedController.UpdateSpeed(speed)
	}); err != nil {
		return err
	}

	// Need to disable BLE notifications when done
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

	err := m.bleDetails.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

		if result.Address.String() == m.bleDetails.bleConfig.SensorUUID {

			// Found the BLE peripheral, stop scanning
			if err := m.bleDetails.bleAdapter.StopScan(); err != nil {
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

	// Calculate time difference between current and last wheel data
	timeDiff := sm.wheelTime - m.lastWheelTime
	if timeDiff == 0 {
		return 0.0
	}

	// Calculate the rev difference between current and last wheel data
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
		return SpeedMeasurement{}, fmt.Errorf("empty data")
	}

	if data[0]&wheelRevFlag == 0 || len(data) < minDataLength {
		return SpeedMeasurement{}, fmt.Errorf("invalid data format or length")
	}

	return SpeedMeasurement{
		wheelRevs: binary.LittleEndian.Uint32(data[1:]),
		wheelTime: binary.LittleEndian.Uint16(data[5:]),
	}, nil
}
