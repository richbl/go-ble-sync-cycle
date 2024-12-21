package ble_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// Constants for test configuration and messages
const (
	// Speed units
	speedUnitsKMH = "kph"
	speedUnitsMPH = "mph"

	// Test identifiers and parameters
	sensorTestUUID       = "test-uuid"
	testTimeout          = 2 * time.Second
	initialScanDelay     = 2 * time.Second
	wheelCircumferenceMM = 2000

	// Error and test case messages
	noBLEAdapterError      = "Skipping test as BLE adapter is not available"
	emptyData              = "empty data"
	invalidFlags           = "invalid flags"
	validDataKPHFirst      = "valid data kph - first reading"
	validDataKPHSubsequent = "valid data kph - subsequent reading"
	validDataMPHFirst      = "valid data mph - first reading"
)

// init initializes the logger for testing
func init() {
	logger.Initialize("debug")
}

// resetBLEData resets the BLE data for testing
func resetBLEData(controller *ble.BLEController) {
	controller.ProcessBLESpeed([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

// waitForScanReset implements a delay before scanning for a BLE peripheral
func waitForScanReset() {
	time.Sleep(initialScanDelay)
}

// createTestController creates test BLE and speed controllers
func createTestController(speedUnits string) (*ble.BLEController, error) {
	// Create test BLE controller
	bleConfig := config.BLEConfig{
		SensorUUID:      sensorTestUUID,
		ScanTimeoutSecs: 10,
	}

	// Create test speed controller
	speedConfig := config.SpeedConfig{
		SpeedUnits:           speedUnits,
		WheelCircumferenceMM: wheelCircumferenceMM,
	}

	return ble.NewBLEController(bleConfig, speedConfig)
}

// controllersIntegrationTest pauses BLE scan and then creates controllers
func controllersIntegrationTest() (*ble.BLEController, error) {
	// Pause to permit BLE adapter to reset
	waitForScanReset()

	// Create test BLE and speed controllers
	controller, err := createTestController(speedUnitsKMH)
	if err != nil {
		return nil, err
	}

	return controller, nil
}

// createTestContextWithTimeout creates a context with a predefined timeout
func createTestContextWithTimeout(t *testing.T) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	return ctx, cancel
}

// setupTestBLEController creates a test BLE controller and handles BLE adapter errors
func setupTestBLEController(t *testing.T) *ble.BLEController {
	controller, err := controllersIntegrationTest()
	if err != nil {
		t.Skip(noBLEAdapterError)
		return nil
	}

	return controller
}

// TestProcessBLESpeed tests the ProcessBLESpeed() function
func TestProcessBLESpeed(t *testing.T) {
	// Define test cases
	tests := []struct {
		name       string
		data       []byte
		speedUnits string
		want       float64
	}{
		{
			name:       emptyData,
			data:       []byte{},
			speedUnits: speedUnitsKMH,
			want:       0.0,
		},
		{
			name:       invalidFlags,
			data:       []byte{0x00},
			speedUnits: speedUnitsKMH,
			want:       0.0,
		},
		{
			name: validDataKPHFirst,
			data: []byte{
				0x01,                   // flags
				0x02, 0x00, 0x00, 0x00, // wheel revs
				0x20, 0x00, // wheel event time
			},
			speedUnits: speedUnitsKMH,
			want:       0.0,
		},
		{
			name: validDataKPHSubsequent,
			data: []byte{
				0x01,                   // flags
				0x03, 0x00, 0x00, 0x00, // wheel revs (1 more revolution)
				0x40, 0x00, // wheel event time (32 time units later)
			},
			speedUnits: speedUnitsKMH,
			want:       225.0, // (1 rev * 2000mm * 3.6) / 32 time units
		},
		{
			name: validDataMPHFirst,
			data: []byte{
				0x01,                   // flags
				0x02, 0x00, 0x00, 0x00, // wheel revs
				0x20, 0x00, // wheel event time
			},
			speedUnits: speedUnitsMPH,
			want:       0.0,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test BLE and speed controllers
			controller, err := createTestController(tt.speedUnits)
			if err != nil {
				t.Skip(noBLEAdapterError)
				return
			}

			// Reset BLE data
			resetBLEData(controller)

			if tt.name == validDataKPHSubsequent {
				controller.ProcessBLESpeed([]byte{
					0x01,                   // flags
					0x02, 0x00, 0x00, 0x00, // initial wheel revs
					0x20, 0x00, // initial wheel time
				})
			}

			// Process BLE data
			got := controller.ProcessBLESpeed(tt.data)
			assert.InDelta(t, tt.want, got, 0.1, "Speed calculation mismatch")
		})
	}

}

// TestNewBLEControllerIntegration tests the creation of a new BLEController
func TestNewBLEControllerIntegration(t *testing.T) {
	// Create test BLE controller
	controller := setupTestBLEController(t)

	assert.NotNil(t, controller)
}

// TestScanForBLEPeripheralIntegration tests the ScanForBLEPeripheral() function
func TestScanForBLEPeripheralIntegration(t *testing.T) {
	// Create test BLE controller
	controller := setupTestBLEController(t)
	ctx, _ := createTestContextWithTimeout(t)

	// Expect error since test UUID won't be found
	_, err := controller.ScanForBLEPeripheral(ctx)

	assert.Error(t, err)
}

// TestGetBLECharacteristicIntegration tests the GetBLECharacteristic() function
func TestGetBLECharacteristicIntegration(t *testing.T) {
	// Create test BLE controller
	controller := setupTestBLEController(t)
	ctx, _ := createTestContextWithTimeout(t)

	// Expect error since test UUID won't be found
	_, err := controller.GetBLECharacteristic(ctx, nil)

	assert.Error(t, err)
}
