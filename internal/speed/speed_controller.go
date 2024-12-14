package speed

import (
	"container/ring"
	"fmt"
	"sync"
	"time"
)

// SpeedController represents the speed controller component
type SpeedController struct {
	speeds        *ring.Ring
	window        int
	currentSpeed  float64
	smoothedSpeed float64
	lastUpdate    time.Time
}

// Mutex for managing concurrent access
var mutex sync.RWMutex

// NewSpeedController creates a new speed controller component which includes a ring buffer for
// storing speed measurements for video playback speed smoothing
func NewSpeedController(window int) *SpeedController {

	// Create ring buffer
	r := ring.New(window)

	// Initialize ring with zero values
	for i := 0; i < window; i++ {
		r.Value = float64(0)
		r = r.Next()
	}

	return &SpeedController{
		speeds: r,
		window: window,
	}
}

// UpdateSpeed updates the current speed measurement and calculates a smoothed average speed over
// the specified window of time
func (t *SpeedController) UpdateSpeed(speed float64) {

	// Lock mutex to prevent concurrent access
	mutex.Lock()
	defer mutex.Unlock()

	// Get speeds
	t.currentSpeed = speed
	t.speeds.Value = speed
	t.speeds = t.speeds.Next()

	// Calculate smoothed speed
	sum := float64(0)

	t.speeds.Do(func(x interface{}) {
		if x != nil {
			sum += x.(float64)
		}
	})

	t.smoothedSpeed = sum / float64(t.window)
	t.lastUpdate = time.Now()
}

// GetSmoothedSpeed returns the smoothed speed measurement
func (t *SpeedController) GetSmoothedSpeed() float64 {

	// Lock mutex to prevent concurrent access
	mutex.RLock()
	defer mutex.RUnlock()

	return t.smoothedSpeed
}

// GetSpeedBuffer returns the speed buffer as an array of string
func (t *SpeedController) GetSpeedBuffer() []string {

	// Lock mutex to prevent concurrent access
	mutex.RLock()
	defer mutex.RUnlock()

	// Create speed buffer
	var speeds []string
	t.speeds.Do(func(x interface{}) {
		if x != nil {
			speeds = append(speeds, fmt.Sprintf("%.2f", x.(float64)))
		}
	})

	return speeds
}
