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

// --- reuse the original mock implementations (slim) ---

type mockServiceDiscoverer struct {
	discoverServicesFunc func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error)
}

func (m *mockServiceDiscoverer) DiscoverServices(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
	if m.discoverServicesFunc != nil {
		return m.discoverServicesFunc(uuids)
	}
	return nil, errServiceDiscoveryFailed
}

type mockCharacteristicDiscoverer struct {
	discoverCharacteristicsFunc func(uuids []bluetooth.UUID) ([]CharacteristicReader, error)
}

func (m *mockCharacteristicDiscoverer) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
	if m.discoverCharacteristicsFunc != nil {
		return m.discoverCharacteristicsFunc(uuids)
	}
	return nil, errCharacteristicsDiscoveryFailed
}

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

// --- helpers to reduce duplication in tests ---

func createTestBLEController(t *testing.T) *Controller {
	t.Helper()
	controller, err := NewBLEController(
		config.BLEConfig{ScanTimeoutSecs: 10},
		config.SpeedConfig{},
	)
	require.NoError(t, err)
	return controller
}

func newMockCharReader(read func(p []byte) (n int, err error), uuid func() bluetooth.UUID) *mockCharacteristicReader {
	return &mockCharacteristicReader{
		readFunc: read,
		uuidFunc: uuid,
	}
}

func newMockCharDiscoverer(fn func(uuids []bluetooth.UUID) ([]CharacteristicReader, error)) *mockCharacteristicDiscoverer {
	return &mockCharacteristicDiscoverer{
		discoverCharacteristicsFunc: fn,
	}
}

// Generic helper to run characteristic-level tests for either battery (read) or CSC (discover only).
// - callFn is the controller method under test (e.g. (*Controller).GetBatteryLevel or (*Controller).GetCSCCharacteristics)
// - expectedUUID is used in assert on what discoverCharacteristics is called with
// - serviceBuilder constructs the mock service that will be passed to the call
// - expectErr & expectErrIs allow verifying expected errors
func runCharTest(t *testing.T, callFn func(*Controller, context.Context, []CharacteristicDiscoverer) error, expectedUUID bluetooth.UUID, serviceBuilder func() CharacteristicDiscoverer, expectErr error, expectErrIs error) {
	t.Helper()
	controller := createTestBLEController(t)

	mockService := serviceBuilder()

	err := callFn(controller, context.Background(), []CharacteristicDiscoverer{mockService})
	if expectErr == nil {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		if expectErrIs != nil {
			assert.ErrorIs(t, err, expectErrIs)
		}
	}
}

// --- service discovery tests (these were already table-driven; keep them) ---

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
		tt := tt
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
		tt := tt
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
		tt := tt
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

// --- characteristic tests for battery (read) ---

func TestGetBatteryLevelSuccess(t *testing.T) {
	const expectedBatteryLevel = 85

	serviceBuilder := func() CharacteristicDiscoverer {
		mockChar := newMockCharReader(func(p []byte) (int, error) {
			require.GreaterOrEqual(t, len(p), 1, "buffer too small")
			p[0] = expectedBatteryLevel
			return 1, nil
		}, func() bluetooth.UUID { return batteryCharacteristicUUID })

		return newMockCharDiscoverer(func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
			assert.Equal(t, []bluetooth.UUID{batteryCharacteristicUUID}, uuids)
			return []CharacteristicReader{mockChar}, nil
		})
	}

	runCharTest(t, (*Controller).GetBatteryLevel, batteryCharacteristicUUID, serviceBuilder, nil, nil)
}

func TestGetBatteryLevelNoCharacteristicsFound(t *testing.T) {
	serviceBuilder := func() CharacteristicDiscoverer {
		return newMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil
		})
	}

	runCharTest(t, (*Controller).GetBatteryLevel, batteryCharacteristicUUID, serviceBuilder, ErrNoBatteryCharacteristics, ErrNoBatteryCharacteristics)
}

func TestGetBatteryLevelReadError(t *testing.T) {
	serviceBuilder := func() CharacteristicDiscoverer {
		mockChar := newMockCharReader(func(_ []byte) (int, error) {
			return 0, errCharReadFailed
		}, func() bluetooth.UUID { return batteryCharacteristicUUID })

		return newMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{mockChar}, nil
		})
	}

	runCharTest(t, (*Controller).GetBatteryLevel, batteryCharacteristicUUID, serviceBuilder, errCharReadFailed, errCharReadFailed)
}

func TestGetBatteryLevelEmptyServicesList(t *testing.T) {
	// direct call with empty services to assert no services provided
	controller := createTestBLEController(t)
	err := controller.GetBatteryLevel(context.Background(), []CharacteristicDiscoverer{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no services provided")
}

// --- characteristic tests for CSC (discover only) ---

func TestGetCSCCharacteristicsSuccess(t *testing.T) {
	serviceBuilder := func() CharacteristicDiscoverer {
		mockChar := newMockCharReader(nil, func() bluetooth.UUID { return cscCharacteristicUUID })
		return newMockCharDiscoverer(func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
			assert.Equal(t, []bluetooth.UUID{cscCharacteristicUUID}, uuids)
			return []CharacteristicReader{mockChar}, nil
		})
	}

	runCharTest(t, (*Controller).GetCSCCharacteristics, cscCharacteristicUUID, serviceBuilder, nil, nil)
}

func TestGetCSCCharacteristicsNoCharacteristicsFound(t *testing.T) {
	serviceBuilder := func() CharacteristicDiscoverer {
		return newMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return []CharacteristicReader{}, nil
		})
	}

	runCharTest(t, (*Controller).GetCSCCharacteristics, cscCharacteristicUUID, serviceBuilder, ErrNoCSCCharacteristics, ErrNoCSCCharacteristics)
}

func TestGetCSCCharacteristicsDiscoveryError(t *testing.T) {
	serviceBuilder := func() CharacteristicDiscoverer {
		return newMockCharDiscoverer(func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
			return nil, errCharacteristicsDiscoveryFailed
		})
	}

	// The production method wraps the err with fmt.Errorf(errFormat, ErrCSCCharDiscovery, err)
	// but we still expect the inner error to be the discovery error via ErrorIs check.
	runCharTest(t, (*Controller).GetCSCCharacteristics, cscCharacteristicUUID, serviceBuilder, errCharacteristicsDiscoveryFailed, errCharacteristicsDiscoveryFailed)
}

func TestGetCSCCharacteristicsEmptyServicesList(t *testing.T) {
	controller := createTestBLEController(t)
	err := controller.GetCSCCharacteristics(context.Background(), []CharacteristicDiscoverer{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoServicesProvided)
	assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
}

// --- small tests that were simple and short ---
// deviceServiceWrapper and serviceConfig error type checks remain the same

func TestDeviceServiceWrapperDiscoverCharacteristics(t *testing.T) {
	t.Run("successful discovery", func(_ *testing.T) {
		var _ CharacteristicDiscoverer = &deviceServiceWrapper{}
	})
}

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
