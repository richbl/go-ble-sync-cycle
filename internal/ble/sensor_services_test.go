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

// charDiscoveryFunc is a type for the Controller methods that discover characteristics
type charDiscoveryFunc func(m *Controller, ctx context.Context, services []CharacteristicDiscoverer) error

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

// --- Service Discovery Tests (Kept as is - already table-driven) ---

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

// --- Characteristic Discovery/Read Consolidated Tests ---

// TestCharacteristicDiscoverySuccess tests successful discovery and/or read of characteristics for both services
func TestCharacteristicDiscoverySuccess(t *testing.T) {
	const expectedBatteryLevel = 85

	tests := []struct {
		name     string
		charFunc charDiscoveryFunc
		charUUID bluetooth.UUID
		readMock func(p []byte) (n int, err error)
	}{
		{
			name:     "Battery Level Success (with read)",
			charFunc: (*Controller).GetBatteryLevel,
			charUUID: batteryCharacteristicUUID,
			readMock: func(p []byte) (n int, err error) {
				require.GreaterOrEqual(t, len(p), 1, "buffer too small")
				p[0] = expectedBatteryLevel
				return 1, nil
			},
		},
		{
			name:     "CSC Characteristics Success (no read)",
			charFunc: (*Controller).GetCSCCharacteristics,
			charUUID: cscCharacteristicUUID,
			readMock: nil, // No read needed for CSC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			mockChar := &mockCharacteristicReader{
				readFunc: tt.readMock,
				uuidFunc: func() bluetooth.UUID {
					return tt.charUUID
				},
			}

			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
					assert.Equal(t, []bluetooth.UUID{tt.charUUID}, uuids)
					return []CharacteristicReader{mockChar}, nil
				},
			}

			err := tt.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})
			assert.NoError(t, err)

			// Check that the characteristic was stored correctly (by UUID)
			if tt.charUUID == batteryCharacteristicUUID {
				// The field is an interface (CharacteristicReader). We assert it's not nil and check its UUID.
				assert.NotNil(t, controller.blePeripheralDetails.batteryCharacteristic, "Battery characteristic not stored")
				assert.Equal(t, tt.charUUID, controller.blePeripheralDetails.batteryCharacteristic.UUID(), "Stored characteristic UUID mismatch")
			} else if tt.charUUID == cscCharacteristicUUID {
				assert.NotNil(t, controller.blePeripheralDetails.bleCharacteristic, "CSC characteristic not stored")
				assert.Equal(t, tt.charUUID, controller.blePeripheralDetails.bleCharacteristic.UUID(), "Stored characteristic UUID mismatch")
			}
		})
	}
}

// TestCharacteristicNoCharacteristicsFound tests the scenario where no characteristics are found
func TestCharacteristicNoCharacteristicsFound(t *testing.T) {
	tests := []struct {
		name        string
		charFunc    charDiscoveryFunc
		charUUID    bluetooth.UUID // Added charUUID for differentiation in assertions
		expectedErr error
	}{
		{
			name:        "Battery Service",
			charFunc:    (*Controller).GetBatteryLevel,
			charUUID:    batteryCharacteristicUUID,
			expectedErr: ErrNoBatteryCharacteristics,
		},
		{
			name:        "CSC Service",
			charFunc:    (*Controller).GetCSCCharacteristics,
			charUUID:    cscCharacteristicUUID,
			expectedErr: ErrNoCSCCharacteristics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			// Mock always returns an empty slice of characteristics
			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
					return []CharacteristicReader{}, nil
				},
			}

			err := tt.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			assert.Error(t, err)

			// assert.ErrorIs works correctly for unwrapped and wrapped errors.
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// TestCharacteristicDiscoveryError tests the scenario where characteristic discovery fails
func TestCharacteristicDiscoveryError(t *testing.T) {
	tests := []struct {
		name        string
		charFunc    charDiscoveryFunc
		charUUID    bluetooth.UUID // Added charUUID for differentiation in assertions
		expectedErr error
	}{
		{
			name:        "Battery Service",
			charFunc:    (*Controller).GetBatteryLevel,
			charUUID:    batteryCharacteristicUUID,
			expectedErr: errCharacteristicsDiscoveryFailed,
		},
		{
			name:        "CSC Service",
			charFunc:    (*Controller).GetCSCCharacteristics,
			charUUID:    cscCharacteristicUUID,
			expectedErr: errCharacteristicsDiscoveryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			// Mock returns a discovery error
			mockService := &mockCharacteristicDiscoverer{
				discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
					return nil, errCharacteristicsDiscoveryFailed
				},
			}

			err := tt.charFunc(controller, context.Background(), []CharacteristicDiscoverer{mockService})

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)

			// Use the charUUID to differentiate the wrapping behavior
			if tt.charUUID == cscCharacteristicUUID {
				// CSC characteristic discovery wraps the error in ErrCSCCharDiscovery
				assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
			}
		})
	}
}

// TestCharacteristicEmptyServicesList tests the scenario where an empty list of services is provided
func TestCharacteristicEmptyServicesList(t *testing.T) {
	tests := []struct {
		name        string
		charFunc    charDiscoveryFunc
		charUUID    bluetooth.UUID // Added charUUID for differentiation in assertions
		expectedErr error
	}{
		{
			name:        "Battery Service",
			charFunc:    (*Controller).GetBatteryLevel,
			charUUID:    batteryCharacteristicUUID,
			expectedErr: ErrNoServicesProvided,
		},
		{
			name:        "CSC Service",
			charFunc:    (*Controller).GetCSCCharacteristics,
			charUUID:    cscCharacteristicUUID,
			expectedErr: ErrNoServicesProvided,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := createTestBLEController(t)

			err := tt.charFunc(controller, context.Background(), []CharacteristicDiscoverer{})

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)

			// Use the charUUID to differentiate the wrapping behavior
			if tt.charUUID == cscCharacteristicUUID {
				// CSC characteristic discovery wraps the error in ErrCSCCharDiscovery
				assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
			}
		})
	}
}

// TestCharacteristicReadError tests the scenario where reading the characteristic fails (unique to GetBatteryLevel)
// This test cannot be easily merged into the table as it's the only one of its kind, and doesn't apply to CSC.
func TestCharacteristicReadError(t *testing.T) {

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

// --- Other Tests (Kept as is) ---

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
