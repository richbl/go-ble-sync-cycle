package ble

import (
	"context"
	"fmt"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	"tinygo.org/x/bluetooth"
)

// CSC service and characteristic UUIDs as defined by Bluetooth SIG
var (
	cscServiceUUID        = bluetooth.New16BitUUID(0x1816)
	cscCharacteristicUUID = bluetooth.New16BitUUID(0x2A5B)
)

// GetCSCServices discovers and returns available CSC services from the BLE peripheral
func (m *Controller) GetCSCServices(ctx context.Context, device ServiceDiscoverer) ([]CharacteristicDiscoverer, error) {

	params := actionParams{
		ctx:        ctx,
		action:     func(found chan<- any, errChan chan<- error) { m.discoverCSCServices(device, found, errChan) },
		logMessage: fmt.Sprintf("discovering CSC service %s", cscServiceUUID.String()),
		stopAction: nil,
	}

	// Discover CSC services
	result, err := m.performBLEAction(params)
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

// discoverCSCServices performs the discovery of CSC services
func (m *Controller) discoverCSCServices(device ServiceDiscoverer, found chan<- any, errChan chan<- error) {

	services, err := device.DiscoverServices([]bluetooth.UUID{cscServiceUUID})
	if err != nil {
		errChan <- err
		return
	}

	if len(services) == 0 {
		errChan <- ErrNoCSCServices
		return
	}

	result := make([]CharacteristicDiscoverer, 0, len(services))

	for _, service := range services {
		result = append(result, &deviceServiceWrapper{service: service})
	}

	found <- result
}

// GetCSCCharacteristics discovers and stores the CSC measurement characteristic from the BLE peripheral
func (m *Controller) GetCSCCharacteristics(ctx context.Context, services []CharacteristicDiscoverer) error {

	params := actionParams{
		ctx: ctx,
		action: func(found chan<- any, errChan chan<- error) {
			m.discoverCSCCharacteristics(services, found, errChan)
		},
		logMessage: fmt.Sprintf("discovering CSC characteristic %s", cscCharacteristicUUID.String()),
		stopAction: nil,
	}

	if _, err := m.performBLEAction(params); err != nil {
		return fmt.Errorf(errFormat, ErrCSCCharDiscovery, err)
	}

	logger.Info(logger.BLE, "found CSC characteristic", m.blePeripheralDetails.bleCharacteristic.UUID().String())

	return nil
}

// discoverCSCCharacteristics performs the discovery of CSC characteristics
func (m *Controller) discoverCSCCharacteristics(services []CharacteristicDiscoverer, found chan<- any, errChan chan<- error) {

	if len(services) == 0 {
		errChan <- ErrNoServicesProvided
		return
	}

	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{cscCharacteristicUUID})
	if err != nil {
		errChan <- err
		return
	}

	if len(characteristics) == 0 {
		errChan <- ErrNoCSCCharacteristics
		return
	}

	m.blePeripheralDetails.bleCharacteristic = characteristics[0]
	found <- characteristics
}
