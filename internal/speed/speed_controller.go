package speed

import (
	"container/ring"
	"fmt"
	"sync"
	"time"
)

// SpeedController manages speed measurements with smoothing
type SpeedController struct {
	speeds        *ring.Ring
	window        int
	currentSpeed  float64
	smoothedSpeed float64
	lastUpdate    time.Time
}

// mutex manages concurrent access to SpeedController
var mutex sync.RWMutex

// NewSpeedController creates a new speed controller with a specified window size
func NewSpeedController(window int) *SpeedController {
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

// GetSmoothedSpeed returns the smoothed speed measurement
func (t *SpeedController) GetSmoothedSpeed() float64 {
	mutex.RLock()
	defer mutex.RUnlock()

	return t.smoothedSpeed
}

// GetSpeedBuffer returns the speed buffer as an array of formatted strings
func (t *SpeedController) GetSpeedBuffer() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	var speeds []string
	t.speeds.Do(func(x interface{}) {

		if x != nil {
			speeds = append(speeds, fmt.Sprintf("%.2f", x.(float64)))
		}

	})

	return speeds
}

// UpdateSpeed updates the current speed measurement and calculates a smoothed average
func (t *SpeedController) UpdateSpeed(speed float64) {
	mutex.Lock()
	defer mutex.Unlock()

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
