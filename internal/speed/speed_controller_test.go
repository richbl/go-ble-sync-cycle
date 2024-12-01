package speed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSpeedController(t *testing.T) {
	window := 5
	controller := NewSpeedController(window)
	assert.NotNil(t, controller)
	assert.Equal(t, window, controller.window)
}

func TestSpeedController_UpdateSpeed(t *testing.T) {
	controller := NewSpeedController(5)
	speed := 10.0
	controller.UpdateSpeed(speed)
	assert.NotNil(t, controller.currentSpeed)
	assert.Equal(t, speed, controller.currentSpeed)
}

func TestSpeedController_GetSmoothedSpeed(t *testing.T) {
	controller := NewSpeedController(5)
	speed := 10.0
	controller.UpdateSpeed(speed)
	smoothedSpeed := controller.GetSmoothedSpeed()
	assert.NotNil(t, smoothedSpeed)
	assert.Equal(t, speed, smoothedSpeed)
}

func TestSpeedController_UpdateSpeed_MultipleTimes(t *testing.T) {
	controller := NewSpeedController(5)
	speeds := []float64{10.0, 20.0, 30.0}
	for _, speed := range speeds {
		controller.UpdateSpeed(speed)
	}
	smoothedSpeed := controller.GetSmoothedSpeed()
	assert.NotNil(t, smoothedSpeed)
	assert.True(t, smoothedSpeed > 0)
}

func TestSpeedController_GetSmoothedSpeed_AfterTime(t *testing.T) {
	controller := NewSpeedController(5)
	speed := 10.0
	controller.UpdateSpeed(speed)
	time.Sleep(100 * time.Millisecond)
	smoothedSpeed := controller.GetSmoothedSpeed()
	assert.NotNil(t, smoothedSpeed)
	assert.Equal(t, speed, smoothedSpeed)
}
