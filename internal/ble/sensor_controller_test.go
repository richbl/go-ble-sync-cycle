package ble

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	config "ble-sync-cycle/internal/configuration"
	"ble-sync-cycle/internal/speed"
)

func TestNewBLEController(t *testing.T) {
	bleConfig := config.BLEConfig{}
	speedConfig := config.SpeedConfig{}

	controller, err := NewBLEController(bleConfig, speedConfig)
	assert.NoError(t, err)
	assert.NotNil(t, controller)
}

func TestBLEController_GetBLECharacteristic(t *testing.T) {
	bleConfig := config.BLEConfig{}
	speedConfig := config.SpeedConfig{}

	controller, err := NewBLEController(bleConfig, speedConfig)
	assert.NoError(t, err)

	ctx := context.Background()
	speedController := speed.NewSpeedController(10)

	char, err := controller.GetBLECharacteristic(ctx, speedController)
	assert.NoError(t, err)
	assert.NotNil(t, char)
}

func TestBLEController_GetBLEUpdates(t *testing.T) {
	bleConfig := config.BLEConfig{}
	speedConfig := config.SpeedConfig{}

	controller, err := NewBLEController(bleConfig, speedConfig)
	assert.NoError(t, err)

	ctx := context.Background()
	speedController := speed.NewSpeedController(10)

	char, err := controller.GetBLECharacteristic(ctx, speedController)
	assert.NoError(t, err)

	err = controller.GetBLEUpdates(ctx, speedController, char)
	assert.NoError(t, err)
}

func TestBLEController_scanForBLEPeripheral(t *testing.T) {
	bleConfig := config.BLEConfig{}
	speedConfig := config.SpeedConfig{}

	controller, err := NewBLEController(bleConfig, speedConfig)
	assert.NoError(t, err)

	ctx := context.Background()

	result, err := controller.scanForBLEPeripheral(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBLEController_processBLESpeed(t *testing.T) {
	bleConfig := config.BLEConfig{}
	speedConfig := config.SpeedConfig{}

	controller, err := NewBLEController(bleConfig, speedConfig)
	assert.NoError(t, err)

	data := []byte{0x01, 0x02, 0x03, 0x04}
	speed := controller.processBLESpeed(data)
	assert.NotNil(t, speed)
}
