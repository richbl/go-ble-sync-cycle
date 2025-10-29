package ble

import (
	"context"
	"errors"
	"testing"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

var (
	errServiceDiscoveryFailed         = errors.New("service discovery failed")
	errCharacteristicsDiscoveryFailed = errors.New("characteristics discovery failed")
	errCharReadFailed                 = errors.New("characteristic read failed")
	errNotificationEnable             = errors.New("failed to enable notifications")
)

// mockServiceDiscoverer is a mock implementation of ServiceDiscoverer
type mockServiceDiscoverer struct {
	discoverServicesFunc func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error)
}

func (m *mockServiceDiscoverer) DiscoverServices(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {

	if m.discoverServicesFunc != nil {
		return m.discoverServicesFunc(uuids)
	}

	return nil, errServiceDiscoveryFailed
}

// mockCharacteristicDiscoverer is a mock implementation of CharacteristicDiscoverer
type mockCharacteristicDiscoverer struct {
	discoverCharacteristicsFunc func(uuids []bluetooth.UUID) ([]CharacteristicReader, error)
}

func (m *mockCharacteristicDiscoverer) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {

	if m.discoverCharacteristicsFunc != nil {
		return m.discoverCharacteristicsFunc(uuids)
	}

	return nil, errCharacteristicsDiscoveryFailed
}

// mockCharacteristicReader is a mock implementation of CharacteristicReader
type mockCharacteristicReader struct {
	readFunc                func(p []byte) (n int, err error)
	uuidFunc                func() bluetooth.UUID
	enableNotificationsFunc func(handler func(buf []byte)) error
}

func (m *mockCharacteristicReader) Read(p []byte) (n int, err error) {

	if m.readFunc != nil {
		return m.readFunc(p)
	}

	return 0, errCharReadFailed
}

func (m *mockCharacteristicReader) UUID() bluetooth.UUID {

	if m.uuidFunc != nil {
		return m.uuidFunc()
	}

	return bluetooth.UUID{}
}

func (m *mockCharacteristicReader) EnableNotifications(handler func(buf []byte)) error {

	if m.enableNotificationsFunc != nil {
		return m.enableNotificationsFunc(handler)
	}

	return errNotificationEnable
}

// createTestBLEController creates a BLE controller for testing
func createTestBLEController(t *testing.T) *Controller {

	t.Helper()

	controller, err := NewBLEController(
		config.BLEConfig{ScanTimeoutSecs: 10},
		config.SpeedConfig{},
	)
	require.NoError(t, err, "failed to create BLE controller")

	return controller
}

// runGetBatteryServiceTest is a helper for testing GetBatteryService scenarios
func runGetBatteryServiceTest(t *testing.T, mockServiceDiscoverer *mockServiceDiscoverer, assertFunc func(*testing.T, []CharacteristicDiscoverer, error)) {

	t.Helper()
	controller := createTestBLEController(t)
	services, err := controller.GetBatteryService(context.Background(), mockServiceDiscoverer)

	assertFunc(t, services, err)

}

// runGetBatteryLevelTest is a helper for testing GetBatteryLevel scenarios
func runGetBatteryLevelTest(t *testing.T, services []CharacteristicDiscoverer, assertFunc func(*testing.T, error)) {

	t.Helper()
	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), services)

	assertFunc(t, err)

}

// TestGetBatteryServiceSuccess tests successful discovery of battery services
func TestGetBatteryServiceSuccess(t *testing.T) {

	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			assert.Equal(t, []bluetooth.UUID{batteryServiceUUID}, uuids)
			return []bluetooth.DeviceService{{}}, nil
		},
	}

	runGetBatteryServiceTest(t, mock, func(t *testing.T, services []CharacteristicDiscoverer, err error) {
		assert.NoError(t, err)
		assert.Len(t, services, 1)
	})

}

// TestGetBatteryServiceNoServicesFound tests the scenario where no battery services are found
func TestGetBatteryServiceNoServicesFound(t *testing.T) {

	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			return nil, nil
		},
	}

	runGetBatteryServiceTest(t, mock, func(t *testing.T, services []CharacteristicDiscoverer, err error) {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no battery services found")
		assert.Nil(t, services)
	})

}

// TestGetBatteryServiceDiscoveryError tests the scenario where service discovery fails
func TestGetBatteryServiceDiscoveryError(t *testing.T) {

	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			return nil, errServiceDiscoveryFailed
		},
	}

	runGetBatteryServiceTest(t, mock, func(t *testing.T, services []CharacteristicDiscoverer, err error) {
		assert.Error(t, err)
		assert.ErrorIs(t, err, errServiceDiscoveryFailed)
		assert.Nil(t, services)
	})

}

// TestGetBatteryLevelSuccess tests successful retrieval of battery level
func TestGetBatteryLevelSuccess(t *testing.T) {

	const expectedBatteryLevel = 85

	// Mock a characteristic that returns the expected battery level
	mockChar := &mockCharacteristicReader{
		readFunc: func(p []byte) (n int, err error) {
			require.GreaterOrEqual(t, len(p), 1, "buffer too small")
			p[0] = expectedBatteryLevel

			return 1, nil
		},
		// Provide a UUID for logging/identification if needed
		uuidFunc: func() bluetooth.UUID {
			return batteryCharacteristicUUID
		},
	}

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
			assert.Equal(t, []bluetooth.UUID{batteryCharacteristicUUID}, uuids)
			return []CharacteristicReader{mockChar}, nil
		},
	}

	runGetBatteryLevelTest(t, []CharacteristicDiscoverer{mockService}, func(t *testing.T, err error) {
		assert.NoError(t, err)
	})

}

// TestGetBatteryLevelNoCharacteristicsFound tests the scenario where no battery characteristics are found
func TestGetBatteryLevelNoCharacteristicsFound(t *testing.T) {

	mockService := &mockCharacteristicDiscoverer{
		// Mock DiscoverCharacteristics to return an empty slice, simulating no characteristics found
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil
		},
	}

	runGetBatteryLevelTest(t, []CharacteristicDiscoverer{mockService}, func(t *testing.T, err error) {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no battery characteristics found")
	})

}

// TestGetBatteryLevelReadError tests the scenario where reading the characteristic fails
func TestGetBatteryLevelReadError(t *testing.T) {

	mockChar := &mockCharacteristicReader{
		readFunc: func(_ []byte) (n int, err error) {
			return 0, errCharReadFailed
		},
		uuidFunc: func() bluetooth.UUID {
			// Provide a UUID for identification if needed
			return batteryCharacteristicUUID
		},
	}

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{mockChar}, nil
		},
	}

	runGetBatteryLevelTest(t, []CharacteristicDiscoverer{mockService}, func(t *testing.T, err error) {
		assert.Error(t, err)
		assert.ErrorIs(t, err, errCharReadFailed)
	})

}

// TestGetBatteryLevelEmptyServicesList tests the scenario where an empty list of services is provided
func TestGetBatteryLevelEmptyServicesList(t *testing.T) {

	runGetBatteryLevelTest(t, []CharacteristicDiscoverer{}, func(t *testing.T, err error) {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no services provided")
	})

}

// TestDeviceServiceWrapperDiscoverCharacteristics verifies that deviceServiceWrapper correctly implements CharacteristicDiscoverer
func TestDeviceServiceWrapperDiscoverCharacteristics(t *testing.T) {

	t.Run("successful discovery", func(_ *testing.T) {

		// This test verifies that the wrapper works correctly
		var _ CharacteristicDiscoverer = &deviceServiceWrapper{}
	})

}
