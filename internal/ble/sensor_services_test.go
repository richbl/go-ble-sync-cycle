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

// serviceTestConfig defines the configuration for testing a BLE service
type serviceTestConfig struct {
	name                   string
	serviceUUID            bluetooth.UUID
	characteristicUUID     bluetooth.UUID
	serviceFunc            func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
	characteristicFunc     func(*Controller, context.Context, []CharacteristicDiscoverer) error
	noServicesError        error
	noCharacteristicsError error
	isReadOperation        bool
}

// getServiceTestConfigs returns test configurations for all supported services
func getServiceTestConfigs() []serviceTestConfig {
	return []serviceTestConfig{
		{
			name:                   "Battery",
			serviceUUID:            batteryServiceUUID,
			characteristicUUID:     batteryCharacteristicUUID,
			serviceFunc:            (*Controller).GetBatteryService,
			characteristicFunc:     (*Controller).GetBatteryLevel,
			noServicesError:        ErrNoBatteryServices,
			noCharacteristicsError: ErrNoBatteryCharacteristics,
			isReadOperation:        true,
		},
		{
			name:                   "CSC",
			serviceUUID:            cscServiceUUID,
			characteristicUUID:     cscCharacteristicUUID,
			serviceFunc:            (*Controller).GetCSCServices,
			characteristicFunc:     (*Controller).GetCSCCharacteristics,
			noServicesError:        ErrNoCSCServices,
			noCharacteristicsError: ErrNoCSCCharacteristics,
			isReadOperation:        false,
		},
	}
}

// TestServiceDiscoverySuccess tests successful discovery of services
func TestServiceDiscoverySuccess(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Service Discovery", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					assert.Equal(t, []bluetooth.UUID{config.serviceUUID}, uuids)
					return []bluetooth.DeviceService{{}}, nil
				},
			}

			services, err := config.serviceFunc(controller, context.Background(), mock)

			assert.NoError(t, err)
			assert.Len(t, services, 1)
		})
	}
}

// TestServiceDiscoveryNoServicesFound tests the scenario where no services are found
func TestServiceDiscoveryNoServicesFound(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Service No Services Found", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					return nil, nil
				},
			}

			services, err := config.serviceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, config.noServicesError)
			assert.Nil(t, services)
		})
	}
}

// TestServiceDiscoveryError tests the scenario where service discovery fails
func TestServiceDiscoveryError(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Service Discovery Error", func(t *testing.T) {
			controller := createTestBLEController(t)

			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					return nil, errServiceDiscoveryFailed
				},
			}

			services, err := config.serviceFunc(controller, context.Background(), mock)

			assert.Error(t, err)
			assert.ErrorIs(t, err, errServiceDiscoveryFailed)
			assert.Nil(t, services)
		})
	}
}

// TestCharacteristicOperationSuccess tests successful characteristic operations
func TestCharacteristicOperationSuccess(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Characteristic Operation Success", func(t *testing.T) {
			controller := createTestBLEController(t)

			var mockChar *mockCharacteristicReader
			if config.isReadOperation {
				mockChar = &mockCharacteristicReader{
					readFunc: func(p []byte) (n int, err error) {
						require.GreaterOrEqual(t, len(p), 1, "buffer too small")
						p[0] = 85 // Example battery level
						return 1, nil
					},
					uuidFunc: func() bluetooth.UUID {
						return config.characteristicUUID
					},
				}
			} else {
				mockChar = &mockCharacteristicReader{
					uuidFunc: func() bluetooth.UUID {
						return config.characteristicUUID
					},
				}
			}

			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
					assert.Equal(t, []bluetooth.UUID{config.characteristicUUID}, uuids)
					return []CharacteristicReader{mockChar}, nil
				},
			}

			err := config.characteristicFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
			assert.NoError(t, err)
		})
	}
}

// TestCharacteristicOperationNoCharacteristicsFound tests the scenario where no characteristics are found
func TestCharacteristicOperationNoCharacteristicsFound(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" No Characteristics Found", func(t *testing.T) {
			controller := createTestBLEController(t)

			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
					assert.Equal(t, []bluetooth.UUID{config.characteristicUUID}, uuids)
					return []CharacteristicReader{}, nil
				},
			}

			err := config.characteristicFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
			assert.Error(t, err)
			assert.ErrorIs(t, err, config.noCharacteristicsError)
		})
	}
}

// TestCharacteristicOperationDiscoveryError tests the scenario where characteristic discovery fails
func TestCharacteristicOperationDiscoveryError(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Characteristic Discovery Error", func(t *testing.T) {
			controller := createTestBLEController(t)

			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
					assert.Equal(t, []bluetooth.UUID{config.characteristicUUID}, uuids)
					return nil, errCharacteristicsDiscoveryFailed
				},
			}

			err := config.characteristicFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
			assert.Error(t, err)
			assert.ErrorIs(t, err, errCharacteristicsDiscoveryFailed)
		})
	}
}

// TestCharacteristicOperationEmptyServicesList tests the scenario where an empty list of services is provided
func TestCharacteristicOperationEmptyServicesList(t *testing.T) {
	for _, config := range getServiceTestConfigs() {
		t.Run(config.name+" Empty Services List", func(t *testing.T) {
			controller := createTestBLEController(t)

			err := config.characteristicFunc(controller, context.Background(), []CharacteristicDiscoverer{})
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrNoServicesProvided)
		})
	}
}

// TestBatteryLevelReadError tests the specific scenario where reading the battery characteristic fails
func TestBatteryLevelReadError(t *testing.T) {
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
		discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
			assert.Equal(t, []bluetooth.UUID{batteryCharacteristicUUID}, uuids)
			return []CharacteristicReader{mockChar}, nil
		},
	}

	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	assert.Error(t, err)
	assert.ErrorIs(t, err, errCharReadFailed)
}

// TestDeviceServiceWrapperDiscoverCharacteristics verifies that deviceServiceWrapper correctly implements CharacteristicDiscoverer
func TestDeviceServiceWrapperDiscoverCharacteristics(t *testing.T) {
	t.Run("successful discovery", func(_ *testing.T) {
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
