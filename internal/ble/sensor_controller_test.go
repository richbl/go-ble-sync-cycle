package ble

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
)

const (
	speedUnitsKMH        = "km/h"
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
		WheelCircumferenceMM: wheelCircumferenceMM,
		SpeedUnits:           speedUnits,
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

	t.Helper()
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
	ctx, cancel := context.WithTimeout(logger.BackgroundCtx, 1*time.Second)
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

// TestScanCancel creates a test BLE controller
func TestScanCancel(t *testing.T) {

	ctrl := setupTestBLEController(t)
	if ctrl == nil {
		return
	}
	ctx, cancel := context.WithTimeout(logger.BackgroundCtx, 500*time.Millisecond)
	defer cancel()

	_, err := ctrl.ScanForBLEPeripheral(ctx)
	assert.ErrorIs(t, err, ErrScanTimeout) // Should cancel early

}

// TestScanForBLEPeripheralWithCancel tests scanning for BLE peripheral with context cancellation
func TestScanForBLEPeripheralWithCancel(t *testing.T) {

	controller := setupTestBLEController(t)
	if controller == nil {
		return
	}
	ctx, cancel := context.WithTimeout(logger.BackgroundCtx, 500*time.Millisecond) // Short for test
	defer cancel()

	_, err := controller.ScanForBLEPeripheral(ctx)
	require.ErrorIs(t, err, ErrScanTimeout, "Should cancel early via ctx")
	// If device present, would connect but timeout forces cancel

}
