package ble

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

var (
	errServiceDiscoveryFailed         = errors.New("service discovery failed")
	errCharacteristicsDiscoveryFailed = errors.New("characteristics discovery failed")
	errCharReadFailed                 = errors.New("characteristic read failed")
	errBufferTooSmall                 = errors.New("buffer too small")
)

// mockServiceDiscoverer is a mock implementation of ServiceDiscoverer
type mockServiceDiscoverer struct {
	discoverServicesFunc func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error)
}

// DiscoverServices mocks the DiscoverServices method
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

// DiscoverCharacteristics mocks the DiscoverCharacteristics method
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

// Read mocks the Read method
func (m *mockCharacteristicReader) Read(p []byte) (int, error) {

	if m.readFunc != nil {
		return m.readFunc(p)
	}

	return 0, errCharReadFailed
}

// UUID mocks the UUID method
func (m *mockCharacteristicReader) UUID() bluetooth.UUID {

	if m.uuidFunc != nil {
		return m.uuidFunc()
	}

	return bluetooth.UUID{}
}

// EnableNotifications mocks the EnableNotifications method
func (m *mockCharacteristicReader) EnableNotifications(handler func(buf []byte)) error {

	if m.enableNotificationsFunc != nil {
		return m.enableNotificationsFunc(handler)
	}

	return ErrNotificationEnable
}

// serviceTestConfig defines the configuration for testing a BLE service type
type serviceTestConfig struct {
	expectedNoSvcErr   error
	expectedNoCharErr  error
	serviceFunc        func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
	charFunc           func(*Controller, context.Context, []CharacteristicDiscoverer) error
	name               string
	serviceUUID        bluetooth.UUID
	characteristicUUID bluetooth.UUID
}

// testConfigs defines all service configurations to test
var testConfigs = []serviceTestConfig{
	{
		name:               "Battery",
		serviceUUID:        batteryServiceUUID,
		characteristicUUID: batteryCharacteristicUUID,
		serviceFunc:        (*Controller).BatteryService,
		charFunc:           (*Controller).BatteryLevel,
		expectedNoSvcErr:   ErrNoBatteryServices,
		expectedNoCharErr:  ErrNoBatteryCharacteristics,
	},
	{
		name:               "CSC",
		serviceUUID:        cscServiceUUID,
		characteristicUUID: cscCharacteristicUUID,
		serviceFunc:        (*Controller).CSCServices,
		charFunc:           (*Controller).CSCCharacteristics,
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
		readFunc: func(p []byte) (int, error) {
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

			services, err := cfg.serviceFunc(controller, context.Background(), mock)

			require.NoError(t, err)
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

			services, err := cfg.serviceFunc(controller, context.Background(), mock)

			require.Error(t, err)
			require.ErrorIs(t, err, cfg.expectedNoSvcErr)
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

			services, err := cfg.serviceFunc(controller, context.Background(), mock)

			require.Error(t, err)
			require.ErrorIs(t, err, errServiceDiscoveryFailed)
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

			err := cfg.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
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

			err := cfg.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			require.Error(t, err)
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

			err := cfg.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			require.Error(t, err)
			assert.ErrorIs(t, err, errCharacteristicsDiscoveryFailed)
		})
	}

}

// TestCharacteristicsEmptyServicesList tests the scenario where an empty list of services is provided
func TestCharacteristicsEmptyServicesList(t *testing.T) {

	for _, cfg := range testConfigs {
		t.Run(cfg.name+" Characteristics Empty Services List", func(t *testing.T) {
			controller := createTestBLEController(t)

			err := cfg.charFunc(controller, context.Background(), []CharacteristicDiscoverer{})

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrNoServicesProvided)
		})
	}

}

// TestBatteryLevelReadError tests the scenario where reading the battery characteristic fails
func TestBatteryLevelReadError(t *testing.T) {

	controller := createTestBLEController(t)

	mockChar := &mockCharacteristicReader{
		readFunc: func(_ []byte) (int, error) {
			return 0, errCharReadFailed
		},
		uuidFunc: func() bluetooth.UUID {
			return batteryCharacteristicUUID
		},
	}

	mockService := createMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
		return []CharacteristicReader{mockChar}, nil
	})

	err := controller.BatteryLevel(context.Background(), []CharacteristicDiscoverer{mockService})
	require.Error(t, err)
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

// TestBatteryServiceWithCancel tests that BatteryService respects context cancellation
func TestBatteryServiceWithCancel(t *testing.T) {

	controller := createTestBLEController(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create a mock that will block until the context is canceled
	mock := &mockServiceDiscoverer{
		discoverServicesFunc: func(_ []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
			<-ctx.Done() // Wait for the context to be done

			return nil, ctx.Err()
		},
	}

	_, err := controller.BatteryService(ctx, mock)
	require.ErrorIs(t, err, ErrScanTimeout)

}
