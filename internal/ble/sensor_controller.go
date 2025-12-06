package ble

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// CharacteristicReader is an interface for a bluetooth peripheral characteristic
type CharacteristicReader interface {
	EnableNotifications(handler func(buf []byte)) error
	Read(p []byte) (n int, err error)
	UUID() bluetooth.UUID
}

// blePeripheralDetails holds details about the BLE peripheral
type blePeripheralDetails struct {
	bleConfig             config.BLEConfig
	bleAdapter            bluetooth.Adapter
	bleCharacteristic     CharacteristicReader
	batteryCharacteristic CharacteristicReader
	batteryLevel          byte
}

// Controller is a central controller for managing the BLE peripheral
type Controller struct {
	blePeripheralDetails blePeripheralDetails
	speedConfig          config.SpeedConfig
}

// actionParams encapsulates parameters for BLE actions
type actionParams struct {
	ctx        context.Context
	action     func(context.Context, chan<- any, chan<- error)
	logMessage string
	stopAction func() error
}

// Mutex for synchronizing adapter access
var AdapterMu sync.Mutex

// Error definitions
var (
	// General BLE errors
	ErrScanTimeout        = errors.New("scanning time limit reached")
	ErrUnsupportedType    = errors.New("unsupported type")
	ErrNoServicesProvided = errors.New("no services provided for characteristic discovery")

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
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*Controller, error) {

	AdapterMu.Lock()
	defer AdapterMu.Unlock()

	bleAdapter := bluetooth.DefaultAdapter

	if err := bleAdapter.Enable(); err != nil {
		return nil, err
	}
	logger.Info(logger.BLE, "created new BLE central controller")

	return &Controller{
		blePeripheralDetails: blePeripheralDetails{
			bleConfig:  bleConfig,
			bleAdapter: *bleAdapter,
		},
		speedConfig: speedConfig,
	}, nil
}

// ScanForBLEPeripheral scans for a BLE peripheral with the specified BD_ADDR
func (m *Controller) ScanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	params := actionParams{
		ctx:        ctx,
		action:     m.scanAction,
		logMessage: fmt.Sprintf("scanning for BLE peripheral BD_ADDR=%s", m.blePeripheralDetails.bleConfig.SensorBDAddr),
		stopAction: m.blePeripheralDetails.bleAdapter.StopScan,
	}

	result, err := m.performBLEAction(params)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrScanTimeout) {
			return bluetooth.ScanResult{}, err
		}

		return bluetooth.ScanResult{}, err
	}

	typedResult, err := assertBLEType(result, bluetooth.ScanResult{})
	if err != nil {
		return bluetooth.ScanResult{}, err
	}
	logger.Info(logger.BLE, "found BLE peripheral", "BD_ADDR", typedResult.Address.String())

	return typedResult, nil
}

// ConnectToBLEPeripheral connects to the specified BLE peripheral
func (m *Controller) ConnectToBLEPeripheral(ctx context.Context, device bluetooth.ScanResult) (bluetooth.Device, error) {

	params := actionParams{
		ctx: ctx,
		action: func(_ context.Context, found chan<- any, errChan chan<- error) {
			m.connectAction(device, found, errChan)
		},
		logMessage: fmt.Sprintf("connecting to BLE peripheral BD_ADDR=%s", device.Address.String()),
		stopAction: nil,
	}

	result, err := m.performBLEAction(params)
	if err != nil {
		return bluetooth.Device{}, err
	}

	typedResult, err := assertBLEType(result, bluetooth.Device{})
	if err != nil {
		return bluetooth.Device{}, err
	}

	logger.Info(logger.BLE, "BLE peripheral device connected")

	return typedResult, nil
}

// BatteryLevel returns the last read battery level (0-100%)
func (m *Controller) BatteryLevel() byte {
	return m.blePeripheralDetails.batteryLevel
}

// assertBLEType casts the result to the specified type
func assertBLEType[T any](result any, target T) (T, error) {

	typedResult, ok := result.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("%w: expected %T, got %T", ErrUnsupportedType, target, result)
	}

	return typedResult, nil
}

// performBLEAction is a wrapper for performing BLE discovery actions
func (m *Controller) performBLEAction(params actionParams) (any, error) {

	scanCtx, cancel := context.WithTimeout(params.ctx, time.Duration(m.blePeripheralDetails.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()
	found := make(chan any, 1)
	errChan := make(chan error, 1)

	go func() {
		logger.Debug(logger.BLE, params.logMessage)
		params.action(scanCtx, found, errChan)
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
func (m *Controller) handleActionTimeout(ctx context.Context, stopAction func() error) (any, error) {

	if stopAction != nil {

		if err := stopAction(); err != nil {
			fmt.Print("\r") // Clear the ^C character from the terminal line
			logger.Error(logger.BLE, fmt.Sprintf("failed to stop action: %v", err))
		}

	}

	// Check if the context was cancelled due to timeout or interrupt
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, fmt.Errorf("%w (%ds)", ErrScanTimeout, m.blePeripheralDetails.bleConfig.ScanTimeoutSecs)
	}

	fmt.Print("\r") // Clear the ^C character from the terminal line
	logger.Info(logger.BLE, "interrupt detected, stopping BLE device setup...")

	return nil, ctx.Err()
}

// scanAction performs the BLE peripheral scan
func (m *Controller) scanAction(ctx context.Context, found chan<- any, errChan chan<- error) {

	foundChan := make(chan bluetooth.ScanResult, 1)

	if err := m.startScanning(foundChan, ctx); err != nil {
		errChan <- err
		return
	}

	found <- <-foundChan

}

// connectAction performs the connection to the BLE peripheral
func (m *Controller) connectAction(device bluetooth.ScanResult, found chan<- any, errChan chan<- error) {

	dev, err := m.blePeripheralDetails.bleAdapter.Connect(device.Address, bluetooth.ConnectionParams{})
	if err != nil {
		errChan <- err
		return
	}

	found <- dev

}

// startScanning starts the BLE peripheral scan
func (m *Controller) startScanning(found chan<- bluetooth.ScanResult, ctx context.Context) error {

	AdapterMu.Lock()
	defer AdapterMu.Unlock()

	err := m.blePeripheralDetails.bleAdapter.Scan(func(_ *bluetooth.Adapter, result bluetooth.ScanResult) {

		select {
		case <-ctx.Done():
			logger.Debug(logger.BLE, "scan canceled via context")
			return
		default:
		}

		if result.Address.String() == m.blePeripheralDetails.bleConfig.SensorBDAddr {

			if stopErr := m.blePeripheralDetails.bleAdapter.StopScan(); stopErr != nil {
				logger.Error(logger.BLE, fmt.Sprintf("failed to stop scan: %v", stopErr))
			}

			select {
			case found <- result:
			default:
				logger.Warn(logger.BLE, "scan results channel full... try restarting the BLE device")
			}

			return
		}

	})

	if err != nil {
		return err
	}

	return nil
}
