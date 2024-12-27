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

// ShutdownContext encapsulates the shutdown context and cancel function
type ShutdownContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// CleanupManager manages cleanup functions
type CleanupManager struct {
	funcs []func()
}

// GoroutineManager manages goroutine synchronization and timeout handling
type GoroutineManager struct {
	wg      *sync.WaitGroup
	timeout time.Duration
}

// ShutdownManager manages the shutdown process
type ShutdownManager struct {
	shutdownCtx ShutdownContext
	routineMgr  *GoroutineManager
	cleanupMgr  CleanupManager
	terminated  chan struct{}
	cleanupOnce sync.Once
}

// ExitHandler handles graceful shutdown on exit
type ExitHandler struct {
	sm *ShutdownManager
}

// NewShutdownManager creates a new ShutdownManager with the specified timeout
func NewShutdownManager(timeout time.Duration) *ShutdownManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShutdownManager{
		shutdownCtx: ShutdownContext{ctx: ctx, cancel: cancel},
		routineMgr:  NewGoroutineManager(timeout),
		cleanupMgr:  CleanupManager{},
		terminated:  make(chan struct{}),
	}
}

// NewGoroutineManager creates a new GoroutineManager with the specified timeout
func NewGoroutineManager(timeout time.Duration) *GoroutineManager {
	return &GoroutineManager{
		wg:      &sync.WaitGroup{},
		timeout: timeout,
	}
}

// NewExitHandler creates a new ExitHandler with the specified ShutdownManager
func NewExitHandler(sm *ShutdownManager) *ExitHandler {
	return &ExitHandler{sm: sm}
}

// Add adds a cleanup function to the CleanupManager
func (cm *CleanupManager) Add(fn func()) {
	cm.funcs = append(cm.funcs, fn)
}

// Execute executes the cleanup functions in reverse order
func (cm *CleanupManager) Execute() {

	for i := len(cm.funcs) - 1; i >= 0; i-- {
		cm.funcs[i]()
	}

}

// Wait blocks until the shutdown sequence is complete
func (sw *GoroutineManager) Wait() {

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), sw.timeout)
	defer timeoutCancel()
	done := make(chan struct{})

	go func() {
		sw.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-timeoutCtx.Done():
		logger.Warn(logger.APP, "shutdown timed out, some goroutines may not have cleaned up properly")
	}

}

// Wait blocks until the shutdown sequence is complete
func (sm *ShutdownManager) Wait() {
	<-sm.terminated
}

// WaitGroup returns the ShutdownManager's WaitGroup for goroutine synchronization
func (sm *ShutdownManager) WaitGroup() *sync.WaitGroup {
	return sm.routineMgr.wg
}

// Future functionality:
//
// AddCleanupFn adds a cleanup function to be executed during shutdown
// Note that cleanup functions are executed in reverse order of registration
// func (sm *ShutdownManager) AddCleanupFn(fn func()) {
// 	sm.cleanupMgr.Add(fn)
// }

// Start monitors for shutdown signals (SIGINT, SIGTERM) and initiates the shutdown sequence
func (sm *ShutdownManager) Start() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		sm.initiateShutdown()
	}()

}

// initiateShutdown coordinates the shutdown sequence
func (sm *ShutdownManager) initiateShutdown() {

	sm.cleanupOnce.Do(func() {
		sm.shutdownCtx.cancel()
		sm.routineMgr.Wait()
		sm.cleanupMgr.Execute()
		close(sm.terminated)
	})

}

// HandleExit waits for the shutdown to complete and then exits the application
func (h *ExitHandler) HandleExit() {

	if h.sm != nil {
		h.sm.Wait()
	}

	waveGoodbye()
}
