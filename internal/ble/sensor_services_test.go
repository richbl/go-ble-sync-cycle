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

// TestServiceDiscoverySuccess tests successful discovery of services (both battery and CSC)
func TestServiceDiscoverySuccess(t *testing.T) {

	tests := []struct {
		name        string
		serviceUUID bluetooth.UUID
		serviceFunc func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
	}{
		{
			name:        "Battery Service Discovery",
			serviceUUID: batteryServiceUUID,
			serviceFunc: (*Controller).GetBatteryService,
		},
		{
			name:        "CSC Service Discovery",
			serviceUUID: cscServiceUUID,
			serviceFunc: (*Controller).GetCSCServices,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					assert.Equal(t, []bluetooth.UUID{tt.serviceUUID}, uuids)
					return []bluetooth.DeviceService{{}}, nil
				},
			}

			services, err := tt.serviceFunc(controller, context.Background(), mock)

			assert.NoError(t, err)
			assert.Len(t, services, 1)
		})
	}

}

// TestServiceDiscoveryNoServicesFound tests the scenario where no services are found
func TestServiceDiscoveryNoServicesFound(t *testing.T) {

	tests := []struct {
		name        string
		serviceUUID bluetooth.UUID
		serviceFunc func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
		expectedErr error
	}{
		{
			name:        "Battery Service No Services Found",
			serviceUUID: batteryServiceUUID,
			serviceFunc: (*Controller).GetBatteryService,
			expectedErr: ErrNoBatteryServices,
		},
		{
			name:        "CSC Service No Services Found",
			serviceUUID: cscServiceUUID,
			serviceFunc: (*Controller).GetCSCServices,
			expectedErr: ErrNoCSCServices,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					return nil, nil
				},
			}

			services, err := tt.serviceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, services)
		})
	}

}

// TestServiceDiscoveryError tests the scenario where service discovery fails
func TestServiceDiscoveryError(t *testing.T) {

	tests := []struct {
		name        string
		serviceUUID bluetooth.UUID
		serviceFunc func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
		expectedErr error
	}{
		{
			name:        "Battery Service Discovery Error",
			serviceUUID: batteryServiceUUID,
			serviceFunc: (*Controller).GetBatteryService,
			expectedErr: errServiceDiscoveryFailed,
		},
		{
			name:        "CSC Service Discovery Error",
			serviceUUID: cscServiceUUID,
			serviceFunc: (*Controller).GetCSCServices,
			expectedErr: errServiceDiscoveryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					return nil, errServiceDiscoveryFailed
				},
			}

			services, err := tt.serviceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, services)
		})
	}

}

// TestGetBatteryLevelSuccess tests successful retrieval of battery level
func TestGetBatteryLevelSuccess(t *testing.T) {

	controller := createTestBLEController(t)

	const expectedBatteryLevel = 85

	mockChar := &mockCharacteristicReader{
		readFunc: func(p []byte) (n int, err error) {
			require.GreaterOrEqual(t, len(p), 1, "buffer too small")
			p[0] = expectedBatteryLevel

			return 1, nil
		},
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

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.NoError(t, err)

}

// TestGetBatteryLevelNoCharacteristicsFound tests the scenario where no battery characteristics are found
func TestGetBatteryLevelNoCharacteristicsFound(t *testing.T) {

	controller := createTestBLEController(t)

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil
		},
	}

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoBatteryCharacteristics)

}

// TestGetBatteryLevelReadError tests the scenario where reading the characteristic fails
func TestGetBatteryLevelReadError(t *testing.T) {

	controller := createTestBLEController(t)

	mockChar := &mockCharacteristicReader{
		readFunc: func(_ []byte) (n int, err error) {
			return 0, errCharReadFailed
		},
		uuidFunc: func() bluetooth.UUID {
			return batteryCharacteristicUUID
		},
	}

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{mockChar}, nil
		},
	}

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, errCharReadFailed)

}

// TestGetBatteryLevelEmptyServicesList tests the scenario where an empty list of services is provided
func TestGetBatteryLevelEmptyServicesList(t *testing.T) {

	controller := createTestBLEController(t)

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no services provided")

}

// TestGetCSCCharacteristicsSuccess tests successful discovery of CSC characteristics
func TestGetCSCCharacteristicsSuccess(t *testing.T) {

	controller := createTestBLEController(t)

	mockChar := &mockCharacteristicReader{
		uuidFunc: func() bluetooth.UUID {
			return cscCharacteristicUUID
		},
	}

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
			assert.Equal(t, []bluetooth.UUID{cscCharacteristicUUID}, uuids)
			return []CharacteristicReader{mockChar}, nil
		},
	}

	err := controller.GetCSCCharacteristics(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.NoError(t, err)

}

// TestGetCSCCharacteristicsNoCharacteristicsFound tests the scenario where no CSC characteristics are found
func TestGetCSCCharacteristicsNoCharacteristicsFound(t *testing.T) {

	controller := createTestBLEController(t)

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil
		},
	}

	err := controller.GetCSCCharacteristics(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoCSCCharacteristics)
	assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())

}

// TestGetCSCCharacteristicsDiscoveryError tests the scenario where CSC characteristic discovery fails
func TestGetCSCCharacteristicsDiscoveryError(t *testing.T) {

	controller := createTestBLEController(t)

	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return nil, errCharacteristicsDiscoveryFailed
		},
	}

	err := controller.GetCSCCharacteristics(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, errCharacteristicsDiscoveryFailed)
	assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())

}

// TestGetCSCCharacteristicsEmptyServicesList tests the scenario where an empty list of services is provided
func TestGetCSCCharacteristicsEmptyServicesList(t *testing.T) {

	controller := createTestBLEController(t)

	err := controller.GetCSCCharacteristics(context.Background(), []CharacteristicDiscoverer{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoServicesProvided)
	assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())

}

// TestDeviceServiceWrapperDiscoverCharacteristics verifies that deviceServiceWrapper correctly implements CharacteristicDiscoverer
func TestDeviceServiceWrapperDiscoverCharacteristics(t *testing.T) {
	t.Run("successful discovery", func(_ *testing.T) {
		// This test verifies that the wrapper works correctly
		var _ CharacteristicDiscoverer = &deviceServiceWrapper{}
	})
}

// TestServiceConfigErrorTypes verifies that battery and CSC services use different error types
func TestServiceConfigErrorTypes(t *testing.T) {

	t.Run("battery service config has battery errors", func(t *testing.T) {
		assert.Equal(t, ErrNoBatteryServices, batteryServiceConfig.errNoServicesFound)
		assert.Equal(t, ErrNoBatteryCharacteristics, batteryServiceConfig.errNoCharacteristicFound)
	})

	t.Run("CSC service config has CSC errors", func(t *testing.T) {
		assert.Equal(t, ErrNoCSCServices, cscServiceConfig.errNoServicesFound)
		assert.Equal(t, ErrNoCSCCharacteristics, cscServiceConfig.errNoCharacteristicFound)
	})

}
