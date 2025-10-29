package ble

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
)

const (
	speedUnitsKMH        = "kph"
	sensorTestBDAddr     = "test-bd-addr"
	initialScanDelay     = 2 * time.Second
	wheelCircumferenceMM = 2000
)

// init is called to set the log level
func init() {
	logger.Initialize("debug")
}

// createTestController creates a test BLE controller
func createTestController(speedUnits string) (*Controller, error) {

	bleConfig := config.BLEConfig{
		SensorBDAddr:    sensorTestBDAddr,
		ScanTimeoutSecs: 10,
	}

	speedConfig := config.SpeedConfig{
		SpeedUnits:           speedUnits,
		WheelCircumferenceMM: wheelCircumferenceMM,
	}

	return NewBLEController(bleConfig, speedConfig)
}

// waitForScanReset waits for the scan to reset
func waitForScanReset() {
	time.Sleep(initialScanDelay)
}

// controllersIntegrationTest creates a test BLE controller
func controllersIntegrationTest() (*Controller, error) {

	waitForScanReset()
	return createTestController(speedUnitsKMH)
}

// setupTestBLEController creates a test BLE controller
func setupTestBLEController(t *testing.T) *Controller {

	controller, err := controllersIntegrationTest()
	if err != nil {
		t.Skip("Skipping test as BLE adapter is not available")
		return nil
	}

	return controller
}

// TestNewBLEControllerIntegration creates a test BLE controller
func TestNewBLEControllerIntegration(t *testing.T) {

	controller := setupTestBLEController(t)
	assert.NotNil(t, controller)
}

// TestScanForBLEPeripheralIntegration creates a test BLE controller
func TestScanForBLEPeripheralIntegration(t *testing.T) {

	controller := setupTestBLEController(t)
	if controller == nil {
		return
	}

	// This test expects a timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := controller.ScanForBLEPeripheral(ctx)
	assert.Error(t, err)
}

// TestBLEControllerMethods creates a test BLE controller
func TestBLEControllerMethods(t *testing.T) {

	controller := setupTestBLEController(t)
	if controller == nil {
		return
	}

	assert.NotNil(t, controller.blePeripheralDetails)
	assert.NotNil(t, controller.speedConfig)
}
