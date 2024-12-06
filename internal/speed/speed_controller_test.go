package speed

import (
	"sync"
	"testing"
	"time"
)

// TestNewSpeedController tests the creation of a new SpeedController
func TestNewSpeedController(t *testing.T) {

	window := 5
	controller := NewSpeedController(window)

	// Verify initial values
	if controller.window != window {
		t.Errorf("Expected window size %d, got %d", window, controller.window)
	}

	if controller.speeds.Len() != window {
		t.Errorf("Expected ring buffer length %d, got %d", window, controller.speeds.Len())
	}

	if controller.smoothedSpeed != 0 {
		t.Errorf("Expected initial smoothed speed to be 0, got %f", controller.smoothedSpeed)
	}

}

// TestUpdateSpeed tests the UpdateSpeed method and verifies smoothed speed calculation
func TestUpdateSpeed(t *testing.T) {

	controller := NewSpeedController(5)
	speeds := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Update speeds in a controlled manner
	for _, speed := range speeds {
		controller.UpdateSpeed(speed)
	}

	expectedSmoothedSpeed := (1.0 + 2.0 + 3.0 + 4.0 + 5.0) / 5.0
	smoothedSpeed := controller.GetSmoothedSpeed()

	if smoothedSpeed != expectedSmoothedSpeed {
		t.Errorf("Expected smoothed speed %f, got %f", expectedSmoothedSpeed, smoothedSpeed)
	}

}

// TestGetSmoothedSpeed tests the GetSmoothedSpeed method
func TestGetSmoothedSpeed(t *testing.T) {

	controller := NewSpeedController(5)
	controller.UpdateSpeed(10.0)

	// Verify initial smoothed speed
	if smoothed := controller.GetSmoothedSpeed(); smoothed != 2.00 {
		t.Errorf("Expected smoothed speed to be %f, got %f", 2.00, smoothed)
	}

	controller.UpdateSpeed(20.0)

	// Verify updated smoothed speed
	if smoothed := controller.GetSmoothedSpeed(); smoothed != 6.00 {
		t.Errorf("Expected smoothed speed to be %f, got %f", 6.00, smoothed)
	}

}

// TestGetSpeedBuffer tests the GetSpeedBuffer method
func TestGetSpeedBuffer(t *testing.T) {

	// Update speeds in a controlled manner
	controller := NewSpeedController(5)
	speeds := []float64{3.5, 2.5, 1.5, 0.0, 0.0}

	for _, speed := range speeds {
		controller.UpdateSpeed(speed)
	}

	expectedBuffer := []string{"3.50", "2.50", "1.50", "0.00", "0.00"}
	buffer := controller.GetSpeedBuffer()

	for i, val := range expectedBuffer {

		if buffer[i] != val {
			t.Errorf("Expected buffer[%d] to be %s, got %s", i, val, buffer[i])
		}

	}

}

// TestConcurrency tests concurrent updates to the SpeedController
func TestConcurrency(t *testing.T) {

	controller := NewSpeedController(5)
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)

		go func(speed float64) {
			defer wg.Done()
			controller.UpdateSpeed(speed)
			time.Sleep(10 * time.Millisecond) // Simulate some processing time
		}(float64(i))

	}

	wg.Wait()
	smoothedSpeed := controller.GetSmoothedSpeed()

	if smoothedSpeed == 0 {
		t.Error("Expected non-zero smoothed speed after concurrent updates")
	}

}
