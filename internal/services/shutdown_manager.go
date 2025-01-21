package shutdownmanager

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// smContext represents the cancellation context for ShutdownManager
type smContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// ShutdownManager represents a shutdown manager that manages a component lifecycle
type ShutdownManager struct {
	context smContext
	wg      sync.WaitGroup
	timeout time.Duration
	errChan chan error
	cleanup []func()
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(timeout time.Duration) *ShutdownManager {

	// Create a context with a timeout
	ctx, cancel := context.WithCancel(context.Background())

	return &ShutdownManager{
		context: smContext{
			ctx:    ctx,
			cancel: cancel,
		},
		timeout: timeout,
		errChan: make(chan error, 1),
	}
}

// Run starts a service and waits for it to complete
func (sm *ShutdownManager) Run(fn func(context.Context) error) {

	// Add a new service to the wait group
	sm.wg.Add(1)

	go func() {
		defer sm.wg.Done()

		// if the context is canceled, signal the error channel and return
		if err := fn(sm.context.ctx); err != nil && !errors.Is(err, context.Canceled) {

			select {
			case sm.errChan <- err:
				sm.context.cancel()
			default:
			}

		}

	}()

}

// AddCleanup adds a cleanup function to the shutdown manager so that they can be executed when
// the shutdown manager is eventually shut down
func (sm *ShutdownManager) AddCleanup(fn func()) {
	sm.cleanup = append(sm.cleanup, fn)
}

// Start starts the shutdown manager and listens for shutdown signals
func (sm *ShutdownManager) Start() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		sm.Shutdown()
	}()

}

// Shutdown shuts down the shutdown manager
func (sm *ShutdownManager) Shutdown() {

	sm.context.cancel()
	done := make(chan struct{})

	go func() {
		sm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(sm.timeout):
		logger.Warn(logger.APP, "shutdown timed out")
	}

	// Execute cleanup functions in reverse order
	for i := len(sm.cleanup) - 1; i >= 0; i-- {
		sm.cleanup[i]()
	}

}

// Context returns the shutdown manager's context
func (sm *ShutdownManager) Context() *context.Context {
	return &sm.context.ctx
}

// Wait waits for the shutdown manager to finish
func (sm *ShutdownManager) Wait() {

	select {
	case <-sm.context.ctx.Done():
		sm.Shutdown()
	case err := <-sm.errChan:
		if err != nil {
			sm.Shutdown()
		}

	}

}

// HandleExit handles an exit signal and shuts down the shutdown manager
func (sm *ShutdownManager) HandleExit() {
	sm.Shutdown()
}
