package speed

import (
	"container/ring"
	"fmt"
	"sync"
	"time"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// state holds the current speed measurement, smoothed speed, and timestamp
type state struct {
	currentSpeed  float64
	smoothedSpeed float64
	timestamp     time.Time
}

// Controller manages speed measurements with smoothing over a specified time window
type Controller struct {
	mu     sync.RWMutex // protects all fields
	speeds *ring.Ring
	window int
	state  state
}

// Common errors
var (
	errUnsupportedType = fmt.Errorf("unsupported type")
)

const (
	errTypeFormat = "%w: got %T"
)

// NewSpeedController creates a new speed controller with a specified window size, which
// determines the number of speed measurements used for smoothing
func NewSpeedController(window int) *Controller {

	r := ring.New(window)

	for range window {
		r.Value = float64(0)
		r = r.Next()
	}

	return &Controller{
		speeds: r,
		window: window,
	}
}

// UpdateSpeed updates the current speed measurement and calculates a smoothed average
func (sc *Controller) UpdateSpeed(speed float64) {

	// Lock the mutex to protect the fields
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.state.currentSpeed = speed
	sc.speeds.Value = speed
	sc.speeds = sc.speeds.Next()

	var sum float64
	sc.speeds.Do(func(x any) {

		value, ok := x.(float64)
		if !ok {
			logger.Error(logger.BLE, fmt.Errorf(errTypeFormat, errUnsupportedType, value))
			return
		}

		sum += value
	})

	// Ahh... smoothness
	sc.state.smoothedSpeed = sum / float64(sc.window)
	sc.state.timestamp = time.Now()
}

// GetSmoothedSpeed returns the current smoothed speed measurement
func (sc *Controller) GetSmoothedSpeed() float64 {

	// Lock the mutex to protect the fields
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.state.smoothedSpeed
}

// GetSpeedBuffer returns the current speed buffer
func (sc *Controller) GetSpeedBuffer() []string {

	// Lock the mutex to protect the fields
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var speeds []string
	sc.speeds.Do(func(x any) {

		if x != nil {
			value, ok := x.(float64)
			if !ok {
				logger.Error(logger.SPEED, fmt.Errorf(errTypeFormat, errUnsupportedType, value))
				return
			}

			speeds = append(speeds, fmt.Sprintf("%.2f", value))
		}

	})

	return speeds
}
