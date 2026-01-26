package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// smContext represents the cancellation context for ShutdownManager
type smContext struct {
	//nolint:containedctx // ShutdownManager owns this context lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// ShutdownManager manages an application lifecycle
type ShutdownManager struct {
	context    smContext
	errChan    chan error
	cleanup    []func()
	wg         sync.WaitGroup
	timeout    time.Duration
	InstanceID int64
}

// Instance counter to distinguish between shutdown manager objects
var shutdownInstanceCounter atomic.Int64

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(timeout time.Duration) *ShutdownManager {

	instanceID := shutdownInstanceCounter.Add(1)

	logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("creating shutdown manager object (id:%04d)...", instanceID))

	// Create a context with a timeout
	ctx, cancel := context.WithCancel(logger.BackgroundCtx)

	logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("created shutdown manager object (id:%04d)", instanceID))

	return &ShutdownManager{
		context: smContext{
			ctx:    ctx,
			cancel: cancel,
		},
		timeout:    timeout,
		InstanceID: instanceID,
		errChan:    make(chan error, 1),
	}
}

// Run starts a service and waits for it to complete
func (sm *ShutdownManager) Run(fn func(context.Context) error) {

	// Run the function in a goroutine managed by the wait group
	sm.wg.Go(func() {

		// if the context is canceled, signal the error channel and return
		if err := fn(sm.context.ctx); err != nil && !errors.Is(err, context.Canceled) {

			select {
			case sm.errChan <- err:
				sm.context.cancel()
			default:
			}

		}

	})

}

// AddCleanup adds a cleanup function to the shutdown manager
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

	logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("shutting down shutdown manager object (id:%04d)...", sm.InstanceID))

	sm.context.cancel()
	done := make(chan struct{})

	go func() {
		sm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("shutdown manager (id:%04d) services stopped", sm.InstanceID))

	case <-time.After(sm.timeout):
		logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("shutdown manager (id:%04d) shutdown timed out", sm.InstanceID))

	}

	// Execute cleanup functions in reverse order
	for i := len(sm.cleanup) - 1; i >= 0; i-- {
		sm.cleanup[i]()
	}

	logger.Debug(logger.BackgroundCtx, logger.APP, fmt.Sprintf("shutdown manager object (id:%04d) shutdown complete", sm.InstanceID))

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
			logger.Error(sm.context.ctx, logger.APP, fmt.Sprintf("service error: %v", err))
			sm.Shutdown()
		}

	}

}

// WaveHello outputs a welcome message
func WaveHello(ctx context.Context) {
	logger.Info(ctx, logger.APP, config.GetFullVersion()+" starting...")
}

// WaveGoodbye outputs a goodbye message and exits the program
func WaveGoodbye(ctx context.Context) {

	logger.ClearCLILine()
	logger.Info(ctx, logger.APP, config.GetFullVersion()+" shutdown complete. Goodbye")
	os.Exit(0)

}
