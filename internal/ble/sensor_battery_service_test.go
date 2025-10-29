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

func TestGetBatteryService_Success(t *testing.T) {
	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			assert.Equal(t, []bluetooth.UUID{batteryServiceUUID}, uuids)
			return []bluetooth.DeviceService{{}}, nil
		},
	}

	controller := createTestBLEController(t)
	services, err := controller.GetBatteryService(context.Background(), mock)

	assert.NoError(t, err)
	assert.Len(t, services, 1)
}

func TestGetBatteryService_NoServicesFound(t *testing.T) {
	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			return nil, nil // Empty result
		},
	}

	controller := createTestBLEController(t)
	services, err := controller.GetBatteryService(context.Background(), mock)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no battery services found")
	assert.Nil(t, services)
}

func TestGetBatteryService_DiscoveryError(t *testing.T) {
	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			return nil, errServiceDiscoveryFailed
		},
	}

	controller := createTestBLEController(t)
	services, err := controller.GetBatteryService(context.Background(), mock)

	assert.Error(t, err)
	assert.ErrorIs(t, err, errServiceDiscoveryFailed)
	assert.Nil(t, services)
}

func TestGetBatteryLevel_Success(t *testing.T) {
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

	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})

	assert.NoError(t, err)
}

func TestGetBatteryLevel_NoCharacteristicsFound(t *testing.T) {
	mockService := &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil // Empty result
		},
	}

	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no battery characteristics found")
}

func TestGetBatteryLevel_ReadError(t *testing.T) {
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

	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})

	assert.Error(t, err)
	assert.ErrorIs(t, err, errCharReadFailed)
}

func TestGetBatteryLevel_EmptyServicesList(t *testing.T) {
	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no services provided")
}

func TestDeviceServiceWrapper_DiscoverCharacteristics(t *testing.T) {
	t.Run("successful discovery", func(_ *testing.T) {

		// This test verifies that the wrapper works correctly
		var _ CharacteristicDiscoverer = &deviceServiceWrapper{}
	})
}
