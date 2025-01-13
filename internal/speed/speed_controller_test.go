package speed

import (
	"sync"
	"testing"
	"time"
)

// testData holds test constants and data
type testData struct {
	window        int
	speeds        []float64
	expectedSpeed float64
	updateCount   int
	sleepDuration time.Duration
}

var td = testData{
	window:        5,
	speeds:        []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	expectedSpeed: 3.0, // average of speeds
	updateCount:   10,
	sleepDuration: 10 * time.Millisecond,
}

// helper function to calculate average of speeds
func calculateAverage(data []float64) float64 {

	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}

	return sum / float64(len(data))
}

// TestNewSpeedController tests the initialization of a new Controller
func TestNewSpeedController(t *testing.T) {

	controller := NewSpeedController(td.window)

	// Verify initialization
	if got := controller.window; got != td.window {
		t.Errorf("window = %d, want %d", got, td.window)
	}

	// Verify speeds buffer
	if got := controller.speeds.Len(); got != td.window {
		t.Errorf("speeds.Len() = %d, want %d", got, td.window)
	}

	// Verify smoothedSpeed
	if got := controller.state.smoothedSpeed; got != 0 {
		t.Errorf("smoothedSpeed = %f, want 0", got)
	}

}

// TestUpdateSpeed tests the UpdateSpeed method of Controller
func TestUpdateSpeed(t *testing.T) {

	controller := NewSpeedController(td.window)

	// Update with test speeds
	for _, speed := range td.speeds {
		controller.UpdateSpeed(speed)
	}

	got := controller.GetSmoothedSpeed()
	want := calculateAverage(td.speeds)

	if got != want {
		t.Errorf("GetSmoothedSpeed() = %f, want %f", got, want)
	}

}

// TestGetSmoothedSpeed tests the GetSmoothedSpeed method of Controller
func TestGetSmoothedSpeed(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		updates  []float64
		expected float64
	}{
		{"single update", []float64{10.0}, 2.0},
		{"multiple updates", []float64{10.0, 20.0}, 6.0},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewSpeedController(td.window)

			for _, speed := range tt.updates {
				controller.UpdateSpeed(speed)
			}

			if got := controller.GetSmoothedSpeed(); got != tt.expected {
				t.Errorf("GetSmoothedSpeed() = %f, want %f", got, tt.expected)
			}

		})
	}

}

// TestGetSpeedBuffer tests the GetSpeedBuffer method of Controller
func TestGetSpeedBuffer(t *testing.T) {

	// Define test cases
	controller := NewSpeedController(td.window)
	speeds := []float64{3.5, 2.5, 1.5, 0.0, 0.0}
	want := []string{"3.50", "2.50", "1.50", "0.00", "0.00"}

	// Update with test speeds
	for _, speed := range speeds {
		controller.UpdateSpeed(speed)
	}

	// Verify buffer
	got := controller.GetSpeedBuffer()

	for i, val := range want {

		if got[i] != val {
			t.Errorf("GetSpeedBuffer()[%d] = %s, want %s", i, got[i], val)
		}

	}

}

// TestConcurrency tests the UpdateSpeed method of Controller
func TestConcurrency(t *testing.T) {

	// Create Controller
	controller := NewSpeedController(td.window)
	var wg sync.WaitGroup

	// Run concurrent updates
	for i := 1; i <= td.updateCount; i++ {
		wg.Add(1)

		go func(speed float64) {
			defer wg.Done()
			controller.UpdateSpeed(speed)
			time.Sleep(td.sleepDuration)
		}(float64(i))

	}

	wg.Wait()

	if got := controller.GetSmoothedSpeed(); got == 0 {
		t.Error("GetSmoothedSpeed() = 0, want non-zero value after concurrent updates")
	}

}
