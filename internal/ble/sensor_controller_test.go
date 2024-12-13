package ble_test

import (
	"context"
	"testing"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/ble"
	config "github.com/richbl/go-ble-sync-cycle/internal/configuration"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
	"github.com/stretchr/testify/assert"
)

const (
	speedUnitsKMH  = "kph"
	speedUnitsMPH  = "mph"
	sensorTestUUID = "test-uuid"
	noBLEnoTest    = "Skipping test as BLE adapter is not available"
)

// init initializes the logger with the debug level
func init() {

	// logger needed for testing of ble controller component
	logger.Initialize("debug")

}

// resetWheelVars is a helper function to reset the package-level wheel variables
func resetWheelVars() {

	// Send empty data to reset speed state
	if controller, _ := ble.NewBLEController(config.BLEConfig{}, config.SpeedConfig{}); controller != nil {
		controller.ProcessBLESpeed([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

}

// waitForScanReset waits for a brief period to allow any ongoing BLE scans to complete
func waitForScanReset() {

	time.Sleep(2 * time.Second)

}

// TestProcessBLESpeed tests the BLE speed processing functionality
func TestProcessBLESpeed(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		data        []byte
		speedUnits  string
		wheelCircMM int
		want        float64
	}{
		{
			name:        "empty data",
			data:        []byte{},
			speedUnits:  speedUnitsKMH,
			wheelCircMM: 2000,
			want:        0.0,
		},
		{
			name:        "invalid flags",
			data:        []byte{0x00},
			speedUnits:  speedUnitsKMH,
			wheelCircMM: 2000,
			want:        0.0,
		},
		{
			name: "valid data kph - first reading",
			data: []byte{
				0x01,                   // flags
				0x02, 0x00, 0x00, 0x00, // wheel revs
				0x20, 0x00, // wheel event time
			},
			speedUnits:  speedUnitsKMH,
			wheelCircMM: 2000,
			want:        0.0, // First reading returns 0
		},
		{
			name: "valid data kph - subsequent reading",
			data: []byte{
				0x01,                   // flags
				0x03, 0x00, 0x00, 0x00, // wheel revs (1 more revolution)
				0x40, 0x00, // wheel event time (32 time units later)
			},
			speedUnits:  speedUnitsKMH,
			wheelCircMM: 2000,
			want:        225.0, // (1 rev * 2000mm * 3.6) / 32 time units
		},
		{
			name: "valid data mph - first reading",
			data: []byte{
				0x01,                   // flags
				0x02, 0x00, 0x00, 0x00, // wheel revs
				0x20, 0x00, // wheel event time
			},
			speedUnits:  speedUnitsMPH,
			wheelCircMM: 2000,
			want:        0.0,
		},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Reset wheel variables before each test
			resetWheelVars()

			// Create a new BLE controller
			bleConfig := config.BLEConfig{
				SensorUUID:      sensorTestUUID,
				ScanTimeoutSecs: 10,
			}

			// Create a new speed controller
			speedConfig := config.SpeedConfig{
				SpeedUnits:           tt.speedUnits,
				WheelCircumferenceMM: tt.wheelCircMM,
			}

			// Create a new BLE controller
			controller, err := ble.NewBLEController(bleConfig, speedConfig)
			if err != nil {
				t.Skip(noBLEnoTest)
				return
			}

			// For subsequent reading tests, first send the initial reading
			if tt.name == "valid data kph - subsequent reading" {
				controller.ProcessBLESpeed([]byte{
					0x01,                   // flags
					0x02, 0x00, 0x00, 0x00, // initial wheel revs
					0x20, 0x00, // initial wheel time
				})
			}

			// Process the BLE speed data
			got := controller.ProcessBLESpeed(tt.data)

			// Verify the calculated speed
			assert.InDelta(t, tt.want, got, 0.1, "Speed calculation mismatch")

		})
	}

}

// TestNewBLEControllerIntegration tests the creation of a new BLE controller
func TestNewBLEControllerIntegration(t *testing.T) {

	waitForScanReset() // Wait before starting test

	// Create a new BLE controller
	bleConfig := config.BLEConfig{
		SensorUUID:      sensorTestUUID,
		ScanTimeoutSecs: 10,
	}

	// Create a new speed controller
	speedConfig := config.SpeedConfig{
		SpeedUnits:           speedUnitsKMH,
		WheelCircumferenceMM: 2000,
	}

	// Create a new BLE controller
	controller, err := ble.NewBLEController(bleConfig, speedConfig)
	if err != nil {
		t.Skip(noBLEnoTest)
		return
	}

	// Verify that the controller is not nil
	assert.NotNil(t, controller)

}

// TestScanForBLEPeripheralIntegration tests the BLE scanning functionality
func TestScanForBLEPeripheralIntegration(t *testing.T) {

	waitForScanReset() // Wait before starting test

	// Create a new BLE controller
	bleConfig := config.BLEConfig{
		SensorUUID:      sensorTestUUID,
		ScanTimeoutSecs: 1,
	}

	// Create a new speed controller
	speedConfig := config.SpeedConfig{
		SpeedUnits:           speedUnitsKMH,
		WheelCircumferenceMM: 2000,
	}

	// Create a new BLE controller
	controller, err := ble.NewBLEController(bleConfig, speedConfig)
	if err != nil {
		t.Skip(noBLEnoTest)
		return
	}

	// Create a context with a timeout of 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Scan for the test UUID
	_, err = controller.ScanForBLEPeripheral(ctx)
	assert.Error(t, err) // Expect error since test UUID won't be found

}

// TestGetBLECharacteristicIntegration tests the BLE characteristic discovery
func TestGetBLECharacteristicIntegration(t *testing.T) {

	waitForScanReset() // Wait before starting test

	// Create a new BLE controller
	bleConfig := config.BLEConfig{
		SensorUUID:      sensorTestUUID,
		ScanTimeoutSecs: 1,
	}

	// Create a new speed controller
	speedConfig := config.SpeedConfig{
		SpeedUnits:           speedUnitsKMH,
		WheelCircumferenceMM: 2000,
	}

	// Create a new BLE controller
	controller, err := ble.NewBLEController(bleConfig, speedConfig)
	if err != nil {
		t.Skip(noBLEnoTest)
		return
	}

	// Create a context with a timeout of 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = controller.GetBLECharacteristic(ctx, nil)
	assert.Error(t, err) // Expect error since test UUID won't be found

}
