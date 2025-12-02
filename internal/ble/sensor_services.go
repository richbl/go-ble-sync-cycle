package ble

import (
	"context"
	"fmt"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
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
	serviceUUID              bluetooth.UUID
	characteristicUUID       bluetooth.UUID
	errNoServicesFound       error
	errNoCharacteristicFound error
}

// deviceServiceWrapper wraps bluetooth.DeviceService to satisfy the CharacteristicDiscoverer interface
type deviceServiceWrapper struct {
	service bluetooth.DeviceService
	config  serviceConfig
}

// charDiscoveryOptions holds options for characteristic discovery
type charDiscoveryOptions struct {
	cfg            serviceConfig
	services       []CharacteristicDiscoverer
	characteristic *CharacteristicReader
	readValue      bool
}

// DiscoverCharacteristics discovers characteristics and returns them as CharacteristicReader interfaces
func (w *deviceServiceWrapper) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {

	bleChars, err := w.service.DiscoverCharacteristics(uuids)
	if err != nil {
		return nil, err
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
func (m *Controller) discoverServices(cfg serviceConfig, device ServiceDiscoverer, found chan<- any, errChan chan<- error) {

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
func (m *Controller) discoverCharacteristics(opts charDiscoveryOptions, found chan<- any, errChan chan<- error) {

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
		found <- characteristics
		return
	}

	// Read the value (single byte: 0-100 for battery level)
	buffer := make([]byte, 1)
	if _, err := characteristics[0].Read(buffer); err != nil {
		errChan <- err
		return
	}

	found <- buffer[0]
}

// executeAction is a helper that creates actionParams and executes a BLE action
func (m *Controller) executeAction(ctx context.Context, logMessage string, action func(context.Context, chan<- any, chan<- error)) (any, error) {

	params := actionParams{
		ctx:        ctx,
		action:     action,
		logMessage: logMessage,
		stopAction: nil,
	}

	return m.performBLEAction(params)
}

// GetBatteryService discovers and returns available battery services from the BLE peripheral
func (m *Controller) GetBatteryService(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	result, err := m.executeAction(
		ctx,
		fmt.Sprintf("discovering battery service %s", batteryServiceConfig.serviceUUID.String()),
		func(_ context.Context, found chan<- any, errChan chan<- error) {
			m.discoverServices(batteryServiceConfig, device, found, errChan)
		},
	)
	if err != nil {
		return nil, err
	}

	typedResult, err := assertBLEType(result, []CharacteristicDiscoverer{})
	if err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "found battery service")

	return typedResult, nil
}

// GetBatteryLevel reads and logs the current battery level (0-100%) from the BLE peripheral
func (m *Controller) GetBatteryLevel(ctx context.Context, services []CharacteristicDiscoverer) error {

	opts := charDiscoveryOptions{
		cfg:            batteryServiceConfig,
		services:       services,
		characteristic: &m.blePeripheralDetails.batteryCharacteristic,
		readValue:      true,
	}

	result, err := m.executeAction(
		ctx,
		fmt.Sprintf("discovering battery characteristic %s", batteryServiceConfig.characteristicUUID.String()),
		func(_ context.Context, found chan<- any, errChan chan<- error) {
			m.discoverCharacteristics(opts, found, errChan)
		},
	)
	if err != nil {
		return err
	}

	batteryLevel, err := assertBLEType(result, byte(0))
	if err != nil {
		return err
	}

	m.blePeripheralDetails.batteryLevel = batteryLevel
	logger.Info(logger.BLE, "found battery characteristic", m.blePeripheralDetails.batteryCharacteristic.UUID().String())
	logger.Info(logger.BLE, "BLE sensor battery level:", fmt.Sprintf("%d%%", batteryLevel))

	return nil
}

// GetCSCServices discovers and returns available CSC services from the BLE peripheral
func (m *Controller) GetCSCServices(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	result, err := m.executeAction(
		ctx,
		fmt.Sprintf("discovering CSC service %s", cscServiceConfig.serviceUUID.String()),
		func(_ context.Context, found chan<- any, errChan chan<- error) {
			m.discoverServices(cscServiceConfig, device, found, errChan)
		},
	)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrCSCServiceDiscovery, err)
	}

	typedResult, err := assertBLEType(result, []CharacteristicDiscoverer{})
	if err != nil {
		return nil, err
	}

	logger.Info(logger.BLE, "found CSC service")

	return typedResult, nil
}

// GetCSCCharacteristics discovers and stores the CSC measurement characteristic from the BLE peripheral
func (m *Controller) GetCSCCharacteristics(ctx context.Context, services []CharacteristicDiscoverer) error {

	opts := charDiscoveryOptions{
		cfg:            cscServiceConfig,
		services:       services,
		characteristic: &m.blePeripheralDetails.bleCharacteristic,
		readValue:      false,
	}
	_, err := m.executeAction(
		ctx,
		fmt.Sprintf("discovering CSC characteristic %s", cscServiceConfig.characteristicUUID.String()),
		func(_ context.Context, found chan<- any, errChan chan<- error) { // Ignore ctx
			m.discoverCharacteristics(opts, found, errChan)
		},
	)

	if err != nil {
		return fmt.Errorf(errFormat, ErrCSCCharDiscovery, err)
	}

	logger.Info(logger.BLE, "found CSC characteristic", m.blePeripheralDetails.bleCharacteristic.UUID().String())

	return nil
}
