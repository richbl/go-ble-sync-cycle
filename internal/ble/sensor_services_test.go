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

type serviceTestParams struct {
	name         string
	serviceUUID  bluetooth.UUID
	funcDiscover func(*Controller, context.Context, ServiceDiscoverer) ([]CharacteristicDiscoverer, error)
	expectedErr  error
	serviceFound bool
}

type charTestParams struct {
	name               string
	characteristicUUID bluetooth.UUID
	funcDiscoverChar   func(*Controller, context.Context, []CharacteristicDiscoverer) error
	setupMock          func(*testing.T) *mockCharacteristicDiscoverer
	expectedErr        error
	errorAssertFunc    func(*testing.T, error)
}

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

func createTestBLEController(t *testing.T) *Controller {
	t.Helper()
	controller, err := NewBLEController(
		config.BLEConfig{ScanTimeoutSecs: 10},
		config.SpeedConfig{},
	)
	require.NoError(t, err, "failed to create BLE controller")

	return controller
}

// ---- Service Discovery Tests ----

func TestServiceDiscovery(t *testing.T) {

	cases := []serviceTestParams{
		{
			name:         "Battery Service Discovery Success",
			serviceUUID:  batteryServiceUUID,
			funcDiscover: (*Controller).GetBatteryService,
			expectedErr:  nil,
			serviceFound: true,
		},
		{
			name:         "CSC Service Discovery Success",
			serviceUUID:  cscServiceUUID,
			funcDiscover: (*Controller).GetCSCServices,
			expectedErr:  nil,
			serviceFound: true,
		},
		{
			name:         "Battery Service No Services Found",
			serviceUUID:  batteryServiceUUID,
			funcDiscover: (*Controller).GetBatteryService,
			expectedErr:  ErrNoBatteryServices,
			serviceFound: false,
		},
		{
			name:         "CSC Service No Services Found",
			serviceUUID:  cscServiceUUID,
			funcDiscover: (*Controller).GetCSCServices,
			expectedErr:  ErrNoCSCServices,
			serviceFound: false,
		},
		{
			name:         "Battery Service Discovery Error",
			serviceUUID:  batteryServiceUUID,
			funcDiscover: (*Controller).GetBatteryService,
			expectedErr:  errServiceDiscoveryFailed,
			serviceFound: false,
		},
		{
			name:         "CSC Service Discovery Error",
			serviceUUID:  cscServiceUUID,
			funcDiscover: (*Controller).GetCSCServices,
			expectedErr:  errServiceDiscoveryFailed,
			serviceFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := createTestBLEController(t)
			mock := &mockServiceDiscoverer{
				discoverServicesFunc: func(uuids []bluetooth.UUID) ([]bluetooth.DeviceService, error) {
					assert.Equal(t, []bluetooth.UUID{tc.serviceUUID}, uuids)

					if errors.Is(tc.expectedErr, errServiceDiscoveryFailed) {
						return nil, errServiceDiscoveryFailed
					}
					if tc.serviceFound {
						return []bluetooth.DeviceService{{}}, nil
					}

					return nil, nil
				},
			}
			services, err := tc.funcDiscover(controller, context.Background(), mock)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, services)
			} else {
				assert.NoError(t, err)
				assert.Len(t, services, 1)
			}
		})
	}
}

// ---- Characteristic Discovery and Read/Write Tests ----

func TestCharacteristicAccess(t *testing.T) {

	const expectedBatteryLevel = 85

	cases := []charTestParams{
		{
			name:               "Battery Level - Success",
			characteristicUUID: batteryCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetBatteryLevel,
			setupMock: func(t *testing.T) *mockCharacteristicDiscoverer {
				mockChar := &mockCharacteristicReader{
					readFunc: func(p []byte) (n int, err error) {
						require.GreaterOrEqual(t, len(p), 1)
						p[0] = expectedBatteryLevel

						return 1, nil
					},
					uuidFunc: func() bluetooth.UUID { return batteryCharacteristicUUID },
				}

				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
						assert.Equal(t, []bluetooth.UUID{batteryCharacteristicUUID}, uuids)
						return []CharacteristicReader{mockChar}, nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name:               "Battery Level - No Characteristics Found",
			characteristicUUID: batteryCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetBatteryLevel,
			setupMock: func(_ *testing.T) *mockCharacteristicDiscoverer {
				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
						return []CharacteristicReader{}, nil
					},
				}
			},
			expectedErr: ErrNoBatteryCharacteristics,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrNoBatteryCharacteristics)
			},
		},
		{
			name:               "Battery Level - Read Error",
			characteristicUUID: batteryCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetBatteryLevel,
			setupMock: func(_ *testing.T) *mockCharacteristicDiscoverer {
				mockChar := &mockCharacteristicReader{
					readFunc: func(_ []byte) (n int, err error) {
						return 0, errCharReadFailed
					},
					uuidFunc: func() bluetooth.UUID { return batteryCharacteristicUUID },
				}

				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
						return []CharacteristicReader{mockChar}, nil
					},
				}
			},
			expectedErr: errCharReadFailed,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, errCharReadFailed)
			},
		},
		{
			name:               "Battery Level - Empty Services List",
			characteristicUUID: batteryCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetBatteryLevel,
			setupMock:          func(_ *testing.T) *mockCharacteristicDiscoverer { return nil },
			expectedErr:        ErrNoServicesProvided,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "no services provided")
			},
		},
		{
			name:               "CSC Characteristic - Success",
			characteristicUUID: cscCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetCSCCharacteristics,
			setupMock: func(t *testing.T) *mockCharacteristicDiscoverer {
				mockChar := &mockCharacteristicReader{
					uuidFunc: func() bluetooth.UUID { return cscCharacteristicUUID },
				}

				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]CharacteristicReader, error) {
						assert.Equal(t, []bluetooth.UUID{cscCharacteristicUUID}, uuids)
						return []CharacteristicReader{mockChar}, nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name:               "CSC Characteristic - No Characteristics Found",
			characteristicUUID: cscCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetCSCCharacteristics,
			setupMock: func(_ *testing.T) *mockCharacteristicDiscoverer {
				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
						return []CharacteristicReader{}, nil
					},
				}
			},
			expectedErr: ErrNoCSCCharacteristics,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrNoCSCCharacteristics)
				assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
			},
		},
		{
			name:               "CSC Characteristic - Discovery Error",
			characteristicUUID: cscCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetCSCCharacteristics,
			setupMock: func(_ *testing.T) *mockCharacteristicDiscoverer {
				return &mockCharacteristicDiscoverer{
					discoverCharacteristicsFunc: func(_ []bluetooth.UUID) ([]CharacteristicReader, error) {
						return nil, errCharacteristicsDiscoveryFailed
					},
				}
			},
			expectedErr: errCharacteristicsDiscoveryFailed,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, errCharacteristicsDiscoveryFailed)
				assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
			},
		},
		{
			name:               "CSC Characteristic - Empty Services List",
			characteristicUUID: cscCharacteristicUUID,
			funcDiscoverChar:   (*Controller).GetCSCCharacteristics,
			setupMock:          func(_ *testing.T) *mockCharacteristicDiscoverer { return nil },
			expectedErr:        ErrNoServicesProvided,
			errorAssertFunc: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrNoServicesProvided)
				assert.Contains(t, err.Error(), ErrCSCCharDiscovery.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := createTestBLEController(t)
			var services []CharacteristicDiscoverer
			if tc.setupMock != nil {
				mock := tc.setupMock(t)
				if mock != nil {
					services = []CharacteristicDiscoverer{mock}
				}
			} else {
				services = []CharacteristicDiscoverer{}
			}
			err := tc.funcDiscoverChar(controller, context.Background(), services)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				if tc.errorAssertFunc != nil {
					tc.errorAssertFunc(t, err)
				} else {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Interface & Config Tests (unchanged, minimal duplication) ----

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
