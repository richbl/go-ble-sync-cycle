package speed

import (
	"sync"
	"testing"
	"time"
)

const defaultWindow = 5

// calculateAverage calculates the average of a slice of float64
func calculateAverage(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range data {
		sum += value
	}

	return sum / float64(len(data))
}

// TestNewSpeedController verifies SpeedController initialization
func TestNewSpeedController(t *testing.T) {

	// Create new SpeedController
	controller := NewSpeedController(defaultWindow)

	if controller.window != defaultWindow {
		t.Errorf("Expected window size %d, got %d", defaultWindow, controller.window)
	}

	if controller.speeds.Len() != defaultWindow {
		t.Errorf("Expected ring buffer length %d, got %d", defaultWindow, controller.speeds.Len())
	}

	if controller.smoothedSpeed != 0 {
		t.Errorf("Expected initial smoothed speed to be 0, got %f", controller.smoothedSpeed)
	}
}

// TestUpdateSpeed checks speed update and smoothing calculation
func TestUpdateSpeed(t *testing.T) {

	// Create new SpeedController
	controller := NewSpeedController(defaultWindow)
	speeds := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	for _, speed := range speeds {
		controller.UpdateSpeed(speed)
	}

	expectedSmoothedSpeed := calculateAverage(speeds)
	if smoothedSpeed := controller.GetSmoothedSpeed(); smoothedSpeed != expectedSmoothedSpeed {
		t.Errorf("Expected smoothed speed %f, got %f", expectedSmoothedSpeed, smoothedSpeed)
	}
}

// TestGetSmoothedSpeed validates smoothed speed retrieval
func TestGetSmoothedSpeed(t *testing.T) {

	// Create new SpeedController
	controller := NewSpeedController(defaultWindow)

	testCases := []struct {
		speed    float64
		expected float64
	}{
		{10.0, 2.0},
		{20.0, 6.0},
	}

	for _, tc := range testCases {
		controller.UpdateSpeed(tc.speed)
		if smoothed := controller.GetSmoothedSpeed(); smoothed != tc.expected {
			t.Errorf("Expected smoothed speed to be %f, got %f", tc.expected, smoothed)
		}
	}
}

// TestGetSpeedBuffer checks speed buffer formatting
func TestGetSpeedBuffer(t *testing.T) {

	// Create new SpeedController
	controller := NewSpeedController(defaultWindow)
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

// TestConcurrency ensures thread-safe speed updates
func TestConcurrency(t *testing.T) {

	// Create new SpeedController
	controller := NewSpeedController(defaultWindow)

	var wg sync.WaitGroup
	numUpdates := 10
	sleepDuration := 10 * time.Millisecond

	for i := 1; i <= numUpdates; i++ {
		wg.Add(1)

		go func(speed float64) {
			defer wg.Done()
			controller.UpdateSpeed(speed)
			time.Sleep(sleepDuration)
		}(float64(i))
	}

	wg.Wait()

	if smoothedSpeed := controller.GetSmoothedSpeed(); smoothedSpeed == 0 {
		t.Error("Expected non-zero smoothed speed after concurrent updates")
	}
}
