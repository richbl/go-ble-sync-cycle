package main

import (
	"context"
)

// ServiceRunner manages individual service goroutines and their lifecycle
type ServiceRunner struct {
	sm          *ShutdownManager
	serviceName string
	errChan     chan error
}

// NewServiceRunner creates a new ServiceRunner with the specified name and ShutdownManager
func NewServiceRunner(sm *ShutdownManager, name string) *ServiceRunner {

	return &ServiceRunner{
		sm:          sm,
		serviceName: name,
		errChan:     make(chan error, 1),
	}
}

// Run executes the provided function in a goroutine and manages its lifecycle, automatically
// handling cleanup and error propagation
func (sr *ServiceRunner) Run(fn func(context.Context) error) {

	sr.sm.WaitGroup().Add(1)

	go func() {
		defer sr.sm.WaitGroup().Done()

		if err := fn(sr.sm.shutdownCtx.ctx); err != nil && err != context.Canceled {
			sr.errChan <- err

			// Initiate shutdown on error
			if sm := sr.sm; sm != nil {
				sm.shutdownCtx.cancel()
			}

		}

		close(sr.errChan)
	}()
}

// Error returns any error that occurred during service execution
func (sr *ServiceRunner) Error() error {

	select {
	case err := <-sr.errChan:
		return err
	default:
		return nil
	}

}
