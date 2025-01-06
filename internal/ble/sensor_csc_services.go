package ble

import (
	"context"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	"tinygo.org/x/bluetooth"
)

// GetBLEServices retrieves CSC services from the BLE peripheral
func (m *BLEController) GetBLEServices(ctx context.Context, device bluetooth.Device) ([]bluetooth.DeviceService, error) {

	params := actionParams{
		ctx:        ctx,
		action:     func(found chan<- interface{}, errChan chan<- error) { m.discoverServicesAction(device, found, errChan) },
		logMessage: "discovering CSC service " + bluetooth.New16BitUUID(0x1816).String(),
		stopAction: nil,
	}

	// Scan for CSC services
	result, err := m.performBLEAction(params)
	if err != nil {
		return nil, err
	}

	typedResult := result.([]bluetooth.DeviceService)
	logger.Info(logger.BLE, "found CSC service", typedResult[0].UUID().String())
	return typedResult, nil
}

// discoverServicesAction performs the discovery of CSC services
func (m *BLEController) discoverServicesAction(device bluetooth.Device, found chan<- interface{}, errChan chan<- error) {

	services, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0x1816)})
	if err != nil {
		errChan <- err
		return
	}

	found <- services
}

// GetBLECharacteristics retrieves CSC characteristics from the BLE peripheral
func (m *BLEController) GetBLECharacteristics(ctx context.Context, services []bluetooth.DeviceService) error {

	params := actionParams{
		ctx: ctx,
		action: func(found chan<- interface{}, errChan chan<- error) {
			m.discoverCharacteristicsAction(services, found, errChan)
		},
		logMessage: "discovering CSC characteristic " + bluetooth.New16BitUUID(0x2A5B).String(),
		stopAction: nil,
	}

	// Scan for CSC characteristics
	_, err := m.performBLEAction(params)

	if err != nil {
		logger.Error(logger.BLE, "CSC characteristics discovery failed:", err.Error())
		return err
	}

	logger.Info(logger.BLE, "found CSC characteristic", m.bleDetails.bleCharacteristic.UUID().String())
	return nil
}

// discoverCharacteristicsAction performs the discovery of CSC characteristics
func (m *BLEController) discoverCharacteristicsAction(services []bluetooth.DeviceService, found chan<- interface{}, errChan chan<- error) {

	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0x2A5B)})
	if err != nil {
		errChan <- err
		return
	}

	m.bleDetails.bleCharacteristic = &characteristics[0]
	found <- characteristics
}
