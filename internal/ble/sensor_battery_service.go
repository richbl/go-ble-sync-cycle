package ble

import (
	"context"
	"fmt"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	"tinygo.org/x/bluetooth"
)

// ServiceDiscoverer provides methods for discovering BLE services on a device
type ServiceDiscoverer interface {
	DiscoverServices(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error)
}

// CharacteristicDiscoverer provides methods for discovering characteristics within a service
type CharacteristicDiscoverer interface {
	DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error)
}

// Battery service and characteristic UUIDs as defined by Bluetooth SIG
var (
	batteryServiceUUID        = bluetooth.New16BitUUID(0x180F)
	batteryCharacteristicUUID = bluetooth.New16BitUUID(0x2A19)
)

// GetBatteryService discovers and returns available battery services from the BLE peripheral
func (m *Controller) GetBatteryService(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	params := actionParams{
		ctx:        ctx,
		action:     func(found chan<- any, errChan chan<- error) { m.discoverBatteryServices(device, found, errChan) },
		logMessage: fmt.Sprintf("discovering battery service %s", batteryServiceUUID.String()),
		stopAction: nil,
	}

	// Discover battery services
	result, err := m.performBLEAction(params)
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

// deviceServiceWrapper wraps bluetooth.DeviceService to satisfy the CharacteristicDiscoverer interface
type deviceServiceWrapper struct {
	service bluetooth.DeviceService
}

// DiscoverCharacteristics discovers characteristics and returns them as CharacteristicReader interfaces
func (w *deviceServiceWrapper) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {

	bleChars, err := w.service.DiscoverCharacteristics(uuids)
	if err != nil {
		return nil, err
	}

	if len(bleChars) == 0 {
		return nil, ErrNoBatteryCharacteristics
	}

	// Convert []bluetooth.DeviceCharacteristic to []CharacteristicReader
	chars := make([]CharacteristicReader, len(bleChars))
	for i := range bleChars {
		chars[i] = &bleChars[i]
	}

	return chars, nil
}

// discoverBatteryServices performs the discovery of battery services
func (m *Controller) discoverBatteryServices(device ServiceDiscoverer, found chan<- any, errChan chan<- error) {

	services, err := device.DiscoverServices([]bluetooth.UUID{batteryServiceUUID})
	if err != nil {
		errChan <- err
		return
	}

	if len(services) == 0 {
		errChan <- ErrNoBatteryServices
		return
	}

	result := make([]CharacteristicDiscoverer, 0, len(services))

	for _, service := range services {
		result = append(result, &deviceServiceWrapper{service: service})
	}

	found <- result
}

// discoverBatteryCharacteristics performs the discovery of battery characteristics
func (m *Controller) discoverBatteryCharacteristics(services []CharacteristicDiscoverer, found chan<- any, errChan chan<- error, read bool) {

	if len(services) == 0 {
		errChan <- ErrNoServicesProvided
		return
	}

	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{batteryCharacteristicUUID})
	if err != nil {
		errChan <- err
		return
	}

	if len(characteristics) == 0 {
		errChan <- ErrNoBatteryCharacteristics
		return
	}

	m.blePeripheralDetails.batteryCharacteristic = characteristics[0]

	if !read {
		found <- characteristics
		return
	}

	// Read the battery level (single byte: 0-100)
	buffer := make([]byte, 1)
	if _, err := m.blePeripheralDetails.batteryCharacteristic.Read(buffer); err != nil {
		errChan <- err
		return
	}

	found <- buffer[0]
}

// GetBatteryLevel reads and logs the current battery level (0-100%) from the BLE peripheral
func (m *Controller) GetBatteryLevel(ctx context.Context, services []CharacteristicDiscoverer) error {

	params := actionParams{
		ctx: ctx,
		action: func(found chan<- any, errChan chan<- error) {
			m.discoverBatteryCharacteristics(services, found, errChan, true)
		},
		logMessage: fmt.Sprintf("discovering battery characteristic %s", batteryCharacteristicUUID.String()),
		stopAction: nil,
	}

	// Discover battery characteristic
	result, err := m.performBLEAction(params)
	if err != nil {
		return err
	}

	batteryLevel, err := assertBLEType(result, byte(0))
	if err != nil {
		return err
	}

	logger.Info(logger.BLE, "found battery characteristic", m.blePeripheralDetails.batteryCharacteristic.UUID().String())
	logger.Info(logger.BLE, "BLE sensor battery level:", fmt.Sprintf("%d%%", batteryLevel))

	return nil
}
