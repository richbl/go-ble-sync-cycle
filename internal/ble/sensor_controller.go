package ble

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"tinygo.org/x/bluetooth"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// CharacteristicReader is an interface for a bluetooth peripheral characteristic
type CharacteristicReader interface {
	EnableNotifications(handler func(buf []byte)) error
	Read(reader []byte) (n int, err error)
	UUID() bluetooth.UUID
}

// blePeripheralDetails holds details about the BLE peripheral
type blePeripheralDetails struct {
	bleAdapter            bluetooth.Adapter
	bleCharacteristic     CharacteristicReader
	batteryCharacteristic CharacteristicReader
	bleConfig             config.BLEConfig
	batteryLevel          byte
}

// Controller is a central controller for managing the BLE peripheral
type Controller struct {
	blePeripheralDetails blePeripheralDetails
	speedConfig          config.SpeedConfig
	InstanceID           int64
}

// actionParams encapsulates parameters for BLE actions
type actionParams[T any] struct {
	action     func(context.Context, chan<- T, chan<- error)
	stopAction func() error
	logMessage string
}

// Mutex for synchronizing adapter access
var AdapterMu sync.Mutex

// Instance counter to distinguish between controller object instances
var bleInstanceCounter atomic.Int64

// Error definitions
var (
	// General BLE errors
	ErrScanTimeout        = errors.New("scanning time limit reached")
	ErrNoServicesProvided = errors.New("no services provided for characteristic discovery")
	ErrTypeMismatch       = errors.New("type mismatch")

	// Battery service/characteristic errors
	ErrNoBatteryServices        = errors.New("no battery services found")
	ErrNoBatteryCharacteristics = errors.New("no battery characteristics found")

	// CSC service/characteristic errors
	ErrCSCServiceDiscovery  = errors.New("CSC service discovery failed")
	ErrCSCCharDiscovery     = errors.New("CSC characteristic discovery failed")
	ErrNoCSCServices        = errors.New("no CSC services found")
	ErrNoCSCCharacteristics = errors.New("no CSC characteristics found")

	// Speed data processing errors
	ErrNoSpeedData        = errors.New("no speed data reported")
	ErrInvalidSpeedData   = errors.New("invalid data format or length")
	ErrNotificationEnable = errors.New("failed to enable BLE notifications")
)

// Format for wrapping errors
const (
	errFormat = "%v: %w"
)

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(ctx context.Context, bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*Controller, error) {

	AdapterMu.Lock()
	defer AdapterMu.Unlock()

	// Increment instance counter
	instanceID := bleInstanceCounter.Add(1)

	logger.Debug(ctx, logger.BLE, fmt.Sprintf("creating BLE controller object (id:%04d)...", instanceID))

	bleAdapter := bluetooth.DefaultAdapter

	if err := bleAdapter.Enable(); err != nil {
		return nil, fmt.Errorf(errFormat, "failed to enable BLE controller", err)
	}

	logger.Debug(ctx, logger.BLE, fmt.Sprintf("created BLE controller object (id:%04d)", instanceID))

	return &Controller{
		blePeripheralDetails: blePeripheralDetails{
			bleConfig:  bleConfig,
			bleAdapter: *bleAdapter,
		},
		speedConfig: speedConfig,
		InstanceID:  instanceID,
	}, nil
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified BD_ADDR
func (m *Controller) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	params := actionParams[bluetooth.ScanResult]{
		action:     m.scanAction,
		logMessage: "scanning for BLE peripheral BD_ADDR=" + m.blePeripheralDetails.bleConfig.SensorBDAddr,
		stopAction: m.blePeripheralDetails.bleAdapter.StopScan,
	}

	result, err := performBLEAction(ctx, m, params)
	if err != nil {
		return bluetooth.ScanResult{}, err
	}

	logger.Info(ctx, logger.BLE, "found BLE peripheral", "BD_ADDR", result.Address.String())

	return result, nil
}

// ConnectToBLEPeripheral connects to the specified BLE peripheral
func (m *Controller) ConnectToBLEPeripheral(ctx context.Context, device bluetooth.ScanResult) (bluetooth.Device, error) {

	params := actionParams[bluetooth.Device]{
		action: func(_ context.Context, found chan<- bluetooth.Device, errChan chan<- error) {
			m.connectAction(device, found, errChan)
		},
		logMessage: "connecting to BLE peripheral BD_ADDR=" + device.Address.String(),
		stopAction: nil,
	}

	result, err := performBLEAction(ctx, m, params)
	if err != nil {
		return bluetooth.Device{}, err
	}

	logger.Info(ctx, logger.BLE, "BLE peripheral device connected")

	return result, nil
}

// BatteryLevelLast returns the last read battery level (0-100%)
func (m *Controller) BatteryLevelLast() byte {
	return m.blePeripheralDetails.batteryLevel
}

// performBLEAction is a wrapper for performing BLE discovery actions
//
//nolint:ireturn // Generic function returning T
func performBLEAction[T any](ctx context.Context, m *Controller, params actionParams[T]) (T, error) {

	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(m.blePeripheralDetails.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()
	found := make(chan T, 1)
	done := make(chan struct{})
	errChan := make(chan error, 1)

	// Start the BLE discovery action
	go func() {
		defer close(done)
		logger.Debug(scanCtx, logger.BLE, params.logMessage)
		params.action(scanCtx, found, errChan)
	}()

	select {

	case result := <-found:
		return result, nil

	case err := <-errChan:
		var zero T

		return zero, err

	case <-scanCtx.Done():
		var zero T
		err := handleActionTimeout(scanCtx, m, params.stopAction)
		logger.Debug(ctx, logger.BLE, "waiting for BLE peripheral disconnect...")

		<-done // Wait for the action to complete

		logger.Debug(ctx, logger.BLE, "BLE peripheral device disconnected")

		return zero, err
	}

}

// handleActionTimeout handles the timeout or cancellation of the BLE action
func handleActionTimeout(ctx context.Context, m *Controller, stopAction func() error) error {

	if stopAction != nil {

		if err := stopAction(); err != nil {
			logger.Error(ctx, logger.BLE, fmt.Sprintf("failed to stop action: %v", err))
		}

	}

	// Check if the context was cancelled due to timeout or interrupt
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w (%ds)", ErrScanTimeout, m.blePeripheralDetails.bleConfig.ScanTimeoutSecs)
	}

	return fmt.Errorf(errFormat, "user interrupt detected", ctx.Err())
}

// scanAction performs the BLE peripheral scan
func (m *Controller) scanAction(ctx context.Context, found chan<- bluetooth.ScanResult, errChan chan<- error) {

	foundChan := make(chan bluetooth.ScanResult, 1)

	if err := m.startScanning(ctx, foundChan); err != nil {
		errChan <- err

		return
	}

	// Wait for the scan to complete
	select {

	case result := <-foundChan:
		found <- result

	default:
	}

}

// connectAction performs the connection to the BLE peripheral
func (m *Controller) connectAction(device bluetooth.ScanResult, found chan<- bluetooth.Device, errChan chan<- error) {

	dev, err := m.blePeripheralDetails.bleAdapter.Connect(device.Address, bluetooth.ConnectionParams{})
	if err != nil {
		errChan <- err

		return
	}

	found <- dev

}

// startScanning starts the BLE peripheral scan and handles device discovery
func (m *Controller) startScanning(ctx context.Context, found chan<- bluetooth.ScanResult) error {

	// Check if already canceled before starting scanning operation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("session stop requested before BLE scan: %w", err)
	}

	AdapterMu.Lock()
	defer AdapterMu.Unlock()

	// Use an atomic flag to ensure we only trigger the device discovery logic once
	var foundOnce atomic.Bool

	err := m.blePeripheralDetails.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

		// Check if context canceled before continuing scanning operation
		if ctx.Err() != nil {
			_ = adapter.StopScan()

			return
		}

		// Address comparison
		if result.Address.String() == m.blePeripheralDetails.bleConfig.SensorBDAddr {

			if foundOnce.CompareAndSwap(false, true) {
				logger.Debug(ctx, logger.BLE, "BLE peripheral found; stopping scan...")
				_ = adapter.StopScan()

				select {
				case found <- result:
					logger.Debug(ctx, logger.BLE, "scan result sent to controller")
				default:
					logger.Warn(ctx, logger.BLE, "controller object no longer listening; scan results ignored")
				}

			}

		}
	})

	if err != nil {
		return fmt.Errorf(errFormat, "unable to start BLE scan", err)
	}

	return nil
}
