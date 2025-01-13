package ble

import (
	"context"
	"fmt"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	"tinygo.org/x/bluetooth"
)

// CSC service and characteristic UUIDs
var (
	cscServiceUUID        = bluetooth.New16BitUUID(0x1816)
	cscCharacteristicUUID = bluetooth.New16BitUUID(0x2A5B)
)

// GetBLEServices retrieves CSC services from the BLE peripheral
func (m *Controller) GetBLEServices(ctx context.Context, device bluetooth.Device) ([]bluetooth.DeviceService, error) {

	params := actionParams{
		ctx:        ctx,
		action:     func(found chan<- interface{}, errChan chan<- error) { m.discoverServicesAction(device, found, errChan) },
		logMessage: fmt.Sprintf("discovering CSC service %s", cscServiceUUID.String()),
		stopAction: nil,
	}

	// Scan for CSC services
	result, err := m.performBLEAction(params)
	if err != nil {
		return nil, err
	}

	// Check the result type
	var typedResult []bluetooth.DeviceService

	typedResult, err = assertBLEType(result, []bluetooth.DeviceService{})
	if err != nil {
		return []bluetooth.DeviceService{}, err
	}

	logger.Info(logger.BLE, "found CSC service", typedResult[0].UUID().String())

	return typedResult, nil
}

// discoverServicesAction performs the discovery of CSC services
func (m *Controller) discoverServicesAction(device bluetooth.Device, found chan<- interface{}, errChan chan<- error) {

	services, err := device.DiscoverServices([]bluetooth.UUID{cscServiceUUID})
	if err != nil {
		errChan <- err
		return
	}

	found <- services
}

// GetBLECharacteristics retrieves CSC characteristics from the BLE peripheral
func (m *Controller) GetBLECharacteristics(ctx context.Context, services []bluetooth.DeviceService) error {

	params := actionParams{
		ctx: ctx,
		action: func(found chan<- interface{}, errChan chan<- error) {
			m.discoverCharacteristicsAction(services, found, errChan)
		},
		logMessage: fmt.Sprintf("discovering CSC characteristic %s", cscCharacteristicUUID.String()),
		stopAction: nil,
	}

	// Scan for CSC characteristics
	_, err := m.performBLEAction(params)

	if err != nil {
		logger.Error(logger.BLE, "CSC characteristics discovery failed:", err.Error())
		return err
	}

	logger.Info(logger.BLE, "found CSC characteristic", m.blePeripheralDetails.bleCharacteristic.UUID().String())

	return nil
}

// discoverCharacteristicsAction performs the discovery of CSC characteristics
func (m *Controller) discoverCharacteristicsAction(services []bluetooth.DeviceService, found chan<- interface{}, errChan chan<- error) {

	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{cscCharacteristicUUID})
	if err != nil {
		errChan <- err
		return
	}

	m.blePeripheralDetails.bleCharacteristic = &characteristics[0]
	found <- characteristics
}
