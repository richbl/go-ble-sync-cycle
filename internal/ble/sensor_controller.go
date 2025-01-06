package ble

import (
	"context"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"

	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// bleDetails holds details about the BLE peripheral
type bleDetails struct {
	bleConfig         config.BLEConfig
	bleAdapter        bluetooth.Adapter
	bleCharacteristic *bluetooth.DeviceCharacteristic
}

// BLEController is a central controller for managing the BLE peripheral
type BLEController struct {
	bleDetails    bleDetails
	speedConfig   config.SpeedConfig
	lastWheelRevs uint32
	lastWheelTime uint16
}

// actionParams encapsulates parameters for BLE actions
type actionParams struct {
	ctx        context.Context
	action     func(chan<- interface{}, chan<- error)
	logMessage string
	stopAction func() error
}

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*BLEController, error) {

	bleAdapter := bluetooth.DefaultAdapter

	if err := bleAdapter.Enable(); err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "created new BLE central controller")

	return &BLEController{
		bleDetails: bleDetails{
			bleConfig:  bleConfig,
			bleAdapter: *bleAdapter,
		},
		speedConfig: speedConfig,
	}, nil
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified UUID
func (m *BLEController) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	params := actionParams{
		ctx:        ctx,
		action:     m.scanAction,
		logMessage: fmt.Sprintf("scanning for BLE peripheral UUID %s", m.bleDetails.bleConfig.SensorUUID),
		stopAction: m.bleDetails.bleAdapter.StopScan,
	}

	// Perform the BLE scan
	result, err := m.performBLEAction(params)
	if err != nil {
		return bluetooth.ScanResult{}, err
	}

	typedResult := result.(bluetooth.ScanResult)
	logger.Info(logger.BLE, "found BLE peripheral", typedResult.Address.String())
	return typedResult, nil
}

// ConnectToBLEPeripheral connects to the specified BLE peripheral
func (m *BLEController) ConnectToBLEPeripheral(ctx context.Context, device bluetooth.ScanResult) (bluetooth.Device, error) {

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

	typedResult := result.(bluetooth.Device)
	logger.Info(logger.BLE, "BLE peripheral device connected")
	return typedResult, nil
}

// performBLEAction is a wrapper for performing BLE discovery actions
func (m *BLEController) performBLEAction(params actionParams) (interface{}, error) {

	// Create a context with a timeout
	scanCtx, cancel := context.WithTimeout(params.ctx, time.Duration(m.bleDetails.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	// Create channels for signaling action completion
	found := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		logger.Debug(logger.BLE, params.logMessage)
		params.action(found, errChan)
	}()

	return m.handleActionCompletion(scanCtx, found, errChan, params.stopAction)
}

// handleActionCompletion handles the completion of the BLE action
func (m *BLEController) handleActionCompletion(ctx context.Context, found <-chan interface{}, errChan <-chan error, stopAction func() error) (interface{}, error) {

	select {
	case result := <-found:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return m.handleActionTimeout(ctx, stopAction)
	}

}

// handleActionTimeout handles the timeout or cancellation of the BLE action
func (m *BLEController) handleActionTimeout(ctx context.Context, stopAction func() error) (interface{}, error) {

	if stopAction != nil {

		if err := stopAction(); err != nil {
			fmt.Print("\r") // Clear the ^C character from the terminal line
			logger.Error(logger.BLE, "failed to stop action:", err.Error())
		}

	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("scanning time limit reached")
	}

	fmt.Print("\r") // Clear the ^C character from the terminal line
	logger.Info(logger.BLE, "user-generated interrupt, stopping BLE device setup...")
	return nil, ctx.Err()
}

// scanAction performs the BLE peripheral scan
func (m *BLEController) scanAction(found chan<- interface{}, errChan chan<- error) {

	foundChan := make(chan bluetooth.ScanResult, 1)

	if err := m.startScanning(foundChan); err != nil {
		errChan <- err
		return
	}

	found <- <-foundChan
}

// connectAction performs the connection to the BLE peripheral
func (m *BLEController) connectAction(device bluetooth.ScanResult, found chan<- interface{}, errChan chan<- error) {

	dev, err := m.bleDetails.bleAdapter.Connect(device.Address, bluetooth.ConnectionParams{})

	if err != nil {
		errChan <- err
		return
	}

	found <- dev
}

func (m *BLEController) startScanning(found chan<- bluetooth.ScanResult) error {

	err := m.bleDetails.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.Address.String() == m.bleDetails.bleConfig.SensorUUID {
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
