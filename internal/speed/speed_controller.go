package speed

import (
	"container/ring"
	"fmt"
	"sync"
	"time"
)

// SpeedState holds the current speed measurement, smoothed speed, and timestamp
type SpeedState struct {
	currentSpeed  float64
	smoothedSpeed float64
	timestamp     time.Time
}

// SpeedController manages speed measurements with smoothing over a specified time window
type SpeedController struct {
	mu         sync.RWMutex // protects all fields
	speeds     *ring.Ring
	window     int
	speedState SpeedState
}

// NewSpeedController creates a new speed controller with a specified window size, which
// determines the number of speed measurements used for smoothing
func NewSpeedController(window int) *SpeedController {

	r := ring.New(window)

	for i := 0; i < window; i++ {
		r.Value = float64(0)
		r = r.Next()
	}

	return &SpeedController{
		speeds: r,
		window: window,
	}
}

// UpdateSpeed updates the current speed measurement and calculates a smoothed average
func (sc *SpeedController) UpdateSpeed(speed float64) {

	// Lock the mutex to protect the fields
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.speedState.currentSpeed = speed
	sc.speeds.Value = speed
	sc.speeds = sc.speeds.Next()

	var sum float64
	sc.speeds.Do(func(x interface{}) {

		if x != nil {
			sum += x.(float64)
		}

	})

	// Ahh... smoothness
	sc.speedState.smoothedSpeed = sum / float64(sc.window)
	sc.speedState.timestamp = time.Now()
}

// GetSmoothedSpeed returns the current smoothed speed measurement
func (sc *SpeedController) GetSmoothedSpeed() float64 {

	// Lock the mutex to protect the fields
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.speedState.smoothedSpeed
}

// GetSpeedBuffer returns the speed buffer as an array of formatted strings
func (sc *SpeedController) GetSpeedBuffer() []string {

	// Lock the mutex to protect the fields
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var speeds []string
	sc.speeds.Do(func(x interface{}) {

		if x != nil {
			speeds = append(speeds, fmt.Sprintf("%.2f", x.(float64)))
		}

	})

	return speeds
}
