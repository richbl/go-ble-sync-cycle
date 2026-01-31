package ble

import (
	"context"
	"fmt"

	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"tinygo.org/x/bluetooth"
)

// Service UUIDs as defined by Bluetooth SIG
var (
	batteryServiceUUID = bluetooth.New16BitUUID(0x180F)
	cscServiceUUID     = bluetooth.New16BitUUID(0x1816)
)

// Characteristic UUIDs as defined by Bluetooth SIG
var (
	batteryCharacteristicUUID = bluetooth.New16BitUUID(0x2A19)
	cscCharacteristicUUID     = bluetooth.New16BitUUID(0x2A5B)
)

// CSC (Cycling Speed & Cadence) service configuration
var cscServiceConfig = serviceConfig{
	serviceUUID:              cscServiceUUID,
	characteristicUUID:       cscCharacteristicUUID,
	errNoServicesFound:       ErrNoCSCServices,
	errNoCharacteristicFound: ErrNoCSCCharacteristics,
}

// Battery service configuration
var batteryServiceConfig = serviceConfig{
	serviceUUID:              batteryServiceUUID,
	characteristicUUID:       batteryCharacteristicUUID,
	errNoServicesFound:       ErrNoBatteryServices,
	errNoCharacteristicFound: ErrNoBatteryCharacteristics,
}

// ServiceDiscoverer provides methods for discovering BLE services on a device
type ServiceDiscoverer interface {
	DiscoverServices(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error)
}

// CharacteristicDiscoverer provides methods for discovering characteristics within a service
type CharacteristicDiscoverer interface {
	DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error)
}

// serviceConfig holds configuration for discovering a specific BLE service
type serviceConfig struct {
	errNoServicesFound       error
	errNoCharacteristicFound error
	serviceUUID              bluetooth.UUID
	characteristicUUID       bluetooth.UUID
}

// deviceServiceWrapper wraps bluetooth.DeviceService to satisfy the CharacteristicDiscoverer interface
type deviceServiceWrapper struct {
	service bluetooth.DeviceService
	config  serviceConfig
}

// charDiscoveryOptions holds options for characteristic discovery
type charDiscoveryOptions struct {
	characteristic *CharacteristicReader
	services       []CharacteristicDiscoverer
	cfg            serviceConfig
	readValue      bool
}

// DiscoverCharacteristics discovers characteristics and returns them as CharacteristicReader interfaces
func (w *deviceServiceWrapper) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {

	bleChars, err := w.service.DiscoverCharacteristics(uuids)

	if err != nil {
		return nil, fmt.Errorf(errFormat, "failed to discover peripheral characteristics", err)
	}

	if len(bleChars) == 0 {
		return nil, w.config.errNoCharacteristicFound
	}

	// Convert []bluetooth.DeviceCharacteristic to []CharacteristicReader
	chars := make([]CharacteristicReader, len(bleChars))

	for i := range bleChars {
		chars[i] = &bleChars[i]
	}

	return chars, nil
}

// discoverServices performs the discovery of BLE services with the given configuration
func discoverServices(cfg serviceConfig, device ServiceDiscoverer, found chan<- []CharacteristicDiscoverer, errChan chan<- error) {

	services, err := device.DiscoverServices([]bluetooth.UUID{cfg.serviceUUID})
	if err != nil {
		errChan <- err

		return
	}

	if len(services) == 0 {
		errChan <- cfg.errNoServicesFound

		return
	}

	result := make([]CharacteristicDiscoverer, 0, len(services))

	for _, service := range services {
		result = append(result, &deviceServiceWrapper{service: service, config: cfg})
	}

	found <- result
}

// discoverCharacteristics performs the discovery of BLE characteristics with the given options
func discoverCharacteristics[T any](opts charDiscoveryOptions, found chan<- T, errChan chan<- error) {

	if len(opts.services) == 0 {
		errChan <- ErrNoServicesProvided

		return
	}

	characteristics, err := opts.services[0].DiscoverCharacteristics([]bluetooth.UUID{opts.cfg.characteristicUUID})
	if err != nil {
		errChan <- err

		return
	}

	if len(characteristics) == 0 {
		errChan <- opts.cfg.errNoCharacteristicFound

		return
	}

	// Store the characteristic if a storage location is provided
	if opts.characteristic != nil {
		*opts.characteristic = characteristics[0]
	}

	// If no read is requested, return the characteristics
	if !opts.readValue {

		if val, ok := any(characteristics).(T); ok {
			found <- val

			return
		}

		errChan <- fmt.Errorf("%w: expected %T, got %T", ErrTypeMismatch, *new(T), characteristics)

		return
	}

	// Read the value (single byte: 0-100 for battery level)
	buffer := make([]byte, 1)
	if _, err := characteristics[0].Read(buffer); err != nil {
		errChan <- err

		return
	}

	if val, ok := any(buffer[0]).(T); ok {
		found <- val

		return
	}

	errChan <- fmt.Errorf("%w: expected byte, got %T", ErrTypeMismatch, buffer[0])
}

// executeAction is a helper that creates actionParams and executes a BLE action
//
//nolint:ireturn // Generic function returning T
func executeAction[T any](ctx context.Context, m *Controller, logMessage string, action func(context.Context, chan<- T, chan<- error)) (T, error) {

	params := actionParams[T]{
		action:     action,
		logMessage: logMessage,
		stopAction: nil,
	}

	return performBLEAction(ctx, m, params)
}

// BatteryService discovers and returns available battery services from the BLE peripheral
func (m *Controller) BatteryService(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	result, err := executeAction(
		ctx,
		m,
		"discovering battery service UUID="+batteryServiceConfig.serviceUUID.String(),
		func(_ context.Context, found chan<- []CharacteristicDiscoverer, errChan chan<- error) {
			discoverServices(batteryServiceConfig, device, found, errChan)
		},
	)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, logger.BLE, "found battery service")

	return result, nil
}

// BatteryLevel reads and logs the current battery level (0-100%) from the BLE peripheral
func (m *Controller) BatteryLevel(ctx context.Context, services []CharacteristicDiscoverer) error {

	opts := charDiscoveryOptions{
		cfg:            batteryServiceConfig,
		services:       services,
		characteristic: &m.blePeripheralDetails.batteryCharacteristic,
		readValue:      true,
	}

	// We explicitly request a byte result here
	batteryLevel, err := executeAction(
		ctx,
		m,
		"discovering battery characteristic UUID="+batteryServiceConfig.characteristicUUID.String(),
		func(_ context.Context, found chan<- byte, errChan chan<- error) {
			discoverCharacteristics(opts, found, errChan)
		},
	)
	if err != nil {
		return err
	}

	m.blePeripheralDetails.batteryLevel = batteryLevel
	logger.Debug(ctx, logger.BLE, "found battery characteristic UUID="+m.blePeripheralDetails.batteryCharacteristic.UUID().String())
	logger.Info(ctx, logger.BLE, fmt.Sprintf("BLE sensor battery level: %d%%", m.blePeripheralDetails.batteryLevel))

	return nil
}

// CSCServices discovers and returns available CSC services from the BLE peripheral
func (m *Controller) CSCServices(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	result, err := executeAction(
		ctx,
		m,
		"discovering CSC service UUID="+cscServiceConfig.serviceUUID.String(),
		func(_ context.Context, found chan<- []CharacteristicDiscoverer, errChan chan<- error) {
			discoverServices(cscServiceConfig, device, found, errChan)
		},
	)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrCSCServiceDiscovery, err)
	}

	logger.Debug(ctx, logger.BLE, "found CSC service UUID="+cscServiceConfig.serviceUUID.String())

	return result, nil
}

// CSCCharacteristics discovers and stores the CSC measurement characteristic from the BLE peripheral
func (m *Controller) CSCCharacteristics(ctx context.Context, services []CharacteristicDiscoverer) error {

	opts := charDiscoveryOptions{
		cfg:            cscServiceConfig,
		services:       services,
		characteristic: &m.blePeripheralDetails.bleCharacteristic,
		readValue:      false,
	}

	// Interested in the CSC measurement characteristic
	_, err := executeAction(
		ctx,
		m,
		"discovering CSC characteristic UUID="+cscServiceConfig.characteristicUUID.String(),
		func(_ context.Context, found chan<- []CharacteristicReader, errChan chan<- error) {
			discoverCharacteristics(opts, found, errChan)
		},
	)

	if err != nil {
		return fmt.Errorf(errFormat, ErrCSCCharDiscovery, err)
	}

	logger.Debug(ctx, logger.BLE, "found CSC characteristic UUID="+cscServiceConfig.characteristicUUID.String())

	return nil
}
