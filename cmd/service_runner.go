package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	logger "github.com/richbl/go-ble-sync-cycle/internal/logging"
)

// ServiceRunner manages individual service goroutines and their lifecycle
type ServiceRunner struct {
	sm          *ShutdownManager
	serviceName string
	errChan     chan error
}

// ShutdownManager handles graceful shutdown of application components and coordinates cleanup
// operations with context cancellations and timeout management
type ShutdownManager struct {
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	cleanupFuncs []func()
	timeout      time.Duration
	terminated   chan struct{}
	cleanupOnce  sync.Once
}

// ExitHandler coordinates the final application shutdown sequence
type ExitHandler struct {
	sm *ShutdownManager
}

// NewShutdownManager creates a new ShutdownManager with the specified timeout duration
func NewShutdownManager(timeout time.Duration) *ShutdownManager {

	ctx, cancel := context.WithCancel(context.Background())

	return &ShutdownManager{
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
		terminated: make(chan struct{}),
		timeout:    timeout,
	}
}

// Context returns the ShutdownManager's context for cancellation propagation
func (sm *ShutdownManager) Context() context.Context {
	return sm.ctx
}

// Wait blocks until the shutdown sequence is complete
func (sm *ShutdownManager) Wait() {
	<-sm.terminated
}

// WaitGroup returns the ShutdownManager's WaitGroup for goroutine synchronization
func (sm *ShutdownManager) WaitGroup() *sync.WaitGroup {
	return &sm.wg
}

// AddCleanupFn adds a cleanup function to be executed during shutdown
// Note that cleanup functions are executed in reverse order of registration
func (sm *ShutdownManager) AddCleanupFn(fn func()) {
	sm.cleanupFuncs = append(sm.cleanupFuncs, fn)
}

// Start begins listening for shutdown signals (SIGINT, SIGTERM)
// When a signal is received, it initiates the shutdown sequence
func (sm *ShutdownManager) Start() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		sm.initiateShutdown()
	}()
}

// initiateShutdown coordinates the shutdown sequence, including timeout management and cleanup
// function execution, and ensures the shutdown sequence runs only once
func (sm *ShutdownManager) initiateShutdown() {

	sm.cleanupOnce.Do(func() {
		sm.cancel()
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), sm.timeout)
		defer timeoutCancel()
		done := make(chan struct{})

		go func() {
			sm.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-timeoutCtx.Done():
			logger.Warn(logger.APP, "shutdown timed out, some goroutines may not have cleaned up properly")
		}

		// Execute cleanup functions (reverse order)
		for i := len(sm.cleanupFuncs) - 1; i >= 0; i-- {
			sm.cleanupFuncs[i]()
		}

		close(sm.terminated)
	})
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

	sr.sm.wg.Add(1)

	go func() {
		defer sr.sm.wg.Done()

		if err := fn(sr.sm.ctx); err != nil && err != context.Canceled {
			sr.errChan <- err
			sr.sm.cancel()
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

// NewExitHandler creates a new ExitHandler with the specified shutdown manager
func NewExitHandler(sm *ShutdownManager) *ExitHandler {
	return &ExitHandler{sm: sm}
}

// HandleExit coordinates the final shutdown sequence and exits the application
func (h *ExitHandler) HandleExit() {

	if h.sm != nil {
		h.sm.Wait()
	}

	waveGoodbye()
}
