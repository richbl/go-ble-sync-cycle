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
	errBufferTooSmall                 = errors.New("buffer too small")
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

// serviceTestConfig defines the configuration for testing a BLE service type
type serviceTestConfig struct {
	name               string
	serviceUUID        bluetooth.UUID
	characteristicUUID bluetooth.UUID
	getServiceFunc     func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
	getCharFunc        func(*Controller, context.Context, []CharacteristicDiscoverer) error
	expectedNoSvcErr   error
	expectedNoCharErr  error
}

// testConfigs defines all service configurations to test
var testConfigs = []serviceTestConfig{
	{
		name:               "Battery",
		serviceUUID:        batteryServiceUUID,
		characteristicUUID: batteryCharacteristicUUID,
		getServiceFunc:     (*Controller).GetBatteryService,
		getCharFunc:        (*Controller).GetBatteryLevel,
		expectedNoSvcErr:   ErrNoBatteryServices,
		expectedNoCharErr:  ErrNoBatteryCharacteristics,
	},
	{
		name:               "CSC",
		serviceUUID:        cscServiceUUID,
		characteristicUUID: cscCharacteristicUUID,
		getServiceFunc:     (*Controller).GetCSCServices,
		getCharFunc:        (*Controller).GetCSCCharacteristics,
		expectedNoSvcErr:   ErrNoCSCServices,
		expectedNoCharErr:  ErrNoCSCCharacteristics,
	},
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

// createMockServiceDiscoverer creates a mock service discoverer with the given function
func createMockServiceDiscoverer(fn func([]bluetooth.UUID) ([]bluetooth.DeviceService, error)) *mockServiceDiscoverer {
	return &mockServiceDiscoverer{discoverServicesFunc: fn}
}

// createMockCharDiscoverer creates a mock characteristic discoverer with the given function
func createMockCharDiscoverer(fn func([]bluetooth.UUID) ([]CharacteristicReader, error)) *mockCharacteristicDiscoverer {
	return &mockCharacteristicDiscoverer{discoverCharacteristicsFunc: fn}
}

// createMockCharReader creates a mock characteristic reader for battery testing
func createMockCharReader(charUUID bluetooth.UUID, batteryLevel byte) *mockCharacteristicReader {
	return &mockCharacteristicReader{
		readFunc: func(p []byte) (n int, err error) {
			if len(p) >= 1 {
				p[0] = batteryLevel
				return 1, nil
			}

			return 0, errBufferTooSmall
		},
		uuidFunc: func() bluetooth.UUID {
			return charUUID
		},
	}
}

// TestServiceDiscoverySuccess tests successful discovery of services
func TestServiceDiscoverySuccess(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Service Discovery", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := createMockServiceDiscoverer(func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
				assert.Equal(t, []bluetooth.UUID{cfg.serviceUUID}, uuids)
				return []bluetooth.DeviceService{{}}, nil
			})

			services, err := cfg.getServiceFunc(controller, context.Background(), mock)

			assert.NoError(t, err)
			assert.Len(t, services, 1)
		})
	}
}

// TestServiceDiscoveryNoServicesFound tests the scenario where no services are found
func TestServiceDiscoveryNoServicesFound(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Service No Services Found", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := createMockServiceDiscoverer(func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
				return nil, nil
			})

			services, err := cfg.getServiceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, cfg.expectedNoSvcErr)
			assert.Nil(t, services)
		})
	}
}

// TestServiceDiscoveryError tests the scenario where service discovery fails
func TestServiceDiscoveryError(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Service Discovery Error", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := createMockServiceDiscoverer(func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
				return nil, errServiceDiscoveryFailed
			})

			services, err := cfg.getServiceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, errServiceDiscoveryFailed)
			assert.Nil(t, services)
		})
	}
}

// TestCharacteristicsDiscoverySuccess tests successful discovery of characteristics
func TestCharacteristicsDiscoverySuccess(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Characteristics Discovery", func(t *testing.T) {
			controller := createTestBLEController(t)

			// Create appropriate mock based on service type
			var mockChar *mockCharacteristicReader
			if cfg.name == "Battery" {
				mockChar = createMockCharReader(cfg.characteristicUUID, 85)
			} else {
				mockChar = &mockCharacteristicReader{
					uuidFunc: func() bluetooth.UUID {
						return cfg.characteristicUUID
					},
				}
			}

			mockService := createMockCharDiscoverer(func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
				assert.Equal(t, []bluetooth.UUID{cfg.characteristicUUID}, uuids)
				return []CharacteristicReader{mockChar}, nil
			})

			err := cfg.getCharFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
			assert.NoError(t, err)
		})
	}
}

// TestCharacteristicsDiscoveryNoCharacteristicsFound tests the scenario where no characteristics are found
func TestCharacteristicsDiscoveryNoCharacteristicsFound(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Characteristics No Characteristics Found", func(t *testing.T) {
			controller := createTestBLEController(t)

			mockService := createMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
				return []CharacteristicReader{}, nil
			})

			err := cfg.getCharFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			assert.Error(t, err)
			assert.ErrorIs(t, err, cfg.expectedNoCharErr)
		})
	}
}

// TestCharacteristicsDiscoveryError tests the scenario where characteristic discovery fails
func TestCharacteristicsDiscoveryError(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Characteristics Discovery Error", func(t *testing.T) {
			controller := createTestBLEController(t)

			mockService := createMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
				return nil, errCharacteristicsDiscoveryFailed
			})

			err := cfg.getCharFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			assert.Error(t, err)
			assert.ErrorIs(t, err, errCharacteristicsDiscoveryFailed)
		})
	}
}

// TestCharacteristicsEmptyServicesList tests the scenario where an empty list of services is provided
func TestCharacteristicsEmptyServicesList(t *testing.T) {
	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Characteristics Empty Services List", func(t *testing.T) {
			controller := createTestBLEController(t)

			err := cfg.getCharFunc(controller, context.Background(), []CharacteristicDiscoverer{})

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrNoServicesProvided)
		})
	}
}

// TestGetBatteryLevelReadError tests the scenario where reading the battery characteristic fails
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

	mockService := createMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
		return []CharacteristicReader{mockChar}, nil
	})

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, errCharReadFailed)
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
