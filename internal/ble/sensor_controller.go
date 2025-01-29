package ble

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// blePeripheralDetails holds details about the BLE peripheral
type blePeripheralDetails struct {
	bleConfig         config.BLEConfig
	bleAdapter        bluetooth.Adapter
	bleCharacteristic *bluetooth.DeviceCharacteristic
}

// Controller is a central controller for managing the BLE peripheral
type Controller struct {
	blePeripheralDetails blePeripheralDetails
	speedConfig          config.SpeedConfig
	lastWheelRevs        uint32
	lastWheelTime        uint16
}

// actionParams encapsulates parameters for BLE actions
type actionParams struct {
	ctx        context.Context
	action     func(chan<- interface{}, chan<- error)
	logMessage string
	stopAction func() error
}

// Error definitions
var (
	errScanTimeout     = fmt.Errorf("scanning time limit reached")
	errUnsupportedType = fmt.Errorf("unsupported type")
)

const (
	errTypeFormat = "%w: got %T"
)

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*Controller, error) {

	bleAdapter := bluetooth.DefaultAdapter

	if err := bleAdapter.Enable(); err != nil {
		logger.Error(logger.BLE, "failed to enable BLE adapter:", err)
		return nil, err
	}

	logger.Info(logger.BLE, "created new BLE central controller")

	return &Controller{
		blePeripheralDetails: blePeripheralDetails{
			bleConfig:         bleConfig,
			bleAdapter:        *bleAdapter,
			bleCharacteristic: &bluetooth.DeviceCharacteristic{},
		},
		lastWheelRevs: 0,
		lastWheelTime: 0,
		speedConfig:   speedConfig,
	}, nil
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified BD_ADDR
func (m *Controller) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	params := actionParams{
		ctx:        ctx,
		action:     m.scanAction,
		logMessage: fmt.Sprintf("scanning for BLE peripheral BD_ADDR %s", m.blePeripheralDetails.bleConfig.SensorBDAddr),
		stopAction: m.blePeripheralDetails.bleAdapter.StopScan,
	}

	// Perform the BLE scan
	result, err := m.performBLEAction(params)
	if err != nil {
		return bluetooth.ScanResult{}, err
	}

	// Check the result type
	var typedResult bluetooth.ScanResult

	typedResult, err = assertBLEType(result, bluetooth.ScanResult{})
	if err != nil {
		return bluetooth.ScanResult{}, err
	}

	logger.Info(logger.BLE, "found BLE peripheral", typedResult.Address.String())

	return typedResult, nil
}

// ConnectToBLEPeripheral connects to the specified BLE peripheral
func (m *Controller) ConnectToBLEPeripheral(ctx context.Context, device bluetooth.ScanResult) (bluetooth.Device, error) {

	params := actionParams{
		ctx:        ctx,
		action:     func(found chan<- interface{}, errChan chan<- error) { m.connectAction(device, found, errChan) },
		logMessage: fmt.Sprintf("connecting to BLE peripheral %s", device.Address.String()),
		stopAction: nil,
	}

	result, err := m.performBLEAction(params)
	if err != nil {
		return bluetooth.Device{}, err
	}

	// Check the result type
	var typedResult bluetooth.Device

	typedResult, err = assertBLEType(result, bluetooth.Device{})
	if err != nil {
		return bluetooth.Device{}, err
	}

	logger.Info(logger.BLE, "BLE peripheral device connected")

	return typedResult, nil
}

// assertBLEType casts the result to the specified type
func assertBLEType[T any](result interface{}, target T) (T, error) {

	typedResult, ok := result.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf(errTypeFormat, errUnsupportedType, target)
	}

	return typedResult, nil
}

// performBLEAction is a wrapper for performing BLE discovery actions
func (m *Controller) performBLEAction(params actionParams) (interface{}, error) {

	// Create a context with a timeout
	scanCtx, cancel := context.WithTimeout(params.ctx, time.Duration(m.blePeripheralDetails.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	// Create channels for signaling action completion
	found := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		logger.Debug(logger.BLE, params.logMessage)
		params.action(found, errChan)
	}()

	select {
	case result := <-found:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-scanCtx.Done():
		return m.handleActionTimeout(scanCtx, params.stopAction)
	}
}

// handleActionTimeout handles the timeout or cancellation of the BLE action
func (m *Controller) handleActionTimeout(ctx context.Context, stopAction func() error) (interface{}, error) {

	if stopAction != nil {

		if err := stopAction(); err != nil {
			fmt.Print("\r") // Clear the ^C character from the terminal line
			logger.Error(logger.BLE, "failed to stop action:", err.Error())
		}

	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, errScanTimeout
	}

	fmt.Print("\r") // Clear the ^C character from the terminal line
	logger.Info(logger.BLE, "interrupt detected, stopping BLE device setup...")

	return nil, ctx.Err()
}

// scanAction performs the BLE peripheral scan
func (m *Controller) scanAction(found chan<- interface{}, errChan chan<- error) {

	foundChan := make(chan bluetooth.ScanResult, 1)

	if err := m.startScanning(foundChan); err != nil {
		errChan <- err
		return
	}

	found <- <-foundChan
}

// connectAction performs the connection to the BLE peripheral
func (m *Controller) connectAction(device bluetooth.ScanResult, found chan<- interface{}, errChan chan<- error) {

	dev, err := m.blePeripheralDetails.bleAdapter.Connect(device.Address, bluetooth.ConnectionParams{})

	if err != nil {
		errChan <- err
		return
	}

	found <- dev
}

// startScanning starts the BLE peripheral scan
func (m *Controller) startScanning(found chan<- bluetooth.ScanResult) error {

	//nolint:revive // bleAdapter.Scan requires adapter struct (though it's not used)
	err := m.blePeripheralDetails.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

		// Stop scanning when the target peripheral is found
		if result.Address.String() == m.blePeripheralDetails.bleConfig.SensorBDAddr {

			if err := m.blePeripheralDetails.bleAdapter.StopScan(); err != nil {
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
