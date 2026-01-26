package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	sm "github.com/richbl/go-ble-sync-cycle/internal/services"
)

// Error messages
var (
	errServiceError = errors.New("service error")
)

// TestNewShutdownManager tests the creation of a new shutdown manager
func TestNewShutdownManager(t *testing.T) {

	logger.Initialize("debug")

	timeout := 5 * time.Second
	manager := sm.NewShutdownManager(timeout)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.Context() == nil {
		t.Fatal("expected non-nil context")
	}

}

// TestRunService tests the Run method of the shutdown manager
func TestRunService(t *testing.T) {

	manager := sm.NewShutdownManager(time.Second)
	serviceDone := make(chan struct{})

	manager.Run(func(ctx context.Context) error {
		defer close(serviceDone)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}

	})

	select {
	case <-serviceDone:
	case <-time.After(2 * time.Second):
		t.Fatal("service did not complete in time")
	}

}

// TestRunServiceError tests the Run method of the shutdown manager with an error
func TestRunServiceError(t *testing.T) {

	manager := sm.NewShutdownManager(time.Second)
	expectedErr := errServiceError

	manager.Run(func(_ context.Context) error {
		return expectedErr
	})

	done := make(chan struct{})
	go func() {
		manager.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Manager detected the error and shut down
	case <-time.After(2 * time.Second):
		t.Fatal("manager did not handle service error")
	}

}

// TestCleanup tests the AddCleanup method of the shutdown manager
func TestCleanup(t *testing.T) {

	manager := sm.NewShutdownManager(time.Second)
	cleanupCalled := false

	manager.AddCleanup(func() {
		cleanupCalled = true
	})

	manager.Shutdown()

	if !cleanupCalled {
		t.Fatal("cleanup function was not called")
	}

}

// TestCleanupOrder tests that cleanup functions are executed in reverse order
func TestCleanupOrder(t *testing.T) {

	manager := sm.NewShutdownManager(time.Second)
	order := make([]int, 0, 3)

	for i := range 3 {
		manager.AddCleanup(func() {
			order = append(order, i)
		})
	}

	manager.Shutdown()

	// Verify reverse order execution
	expected := []int{2, 1, 0}

	for i, v := range order {

		if v != expected[i] {
			t.Errorf("cleanup order incorrect, got %v, want %v", order, expected)
		}

	}

}

// TestShutdown tests the Shutdown method of the shutdown manager
func TestShutdownTimeout(t *testing.T) {

	timeout := 100 * time.Millisecond
	manager := sm.NewShutdownManager(timeout)
	started := make(chan struct{})

	manager.Run(func(ctx context.Context) error {
		close(started)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(timeout * 3):
			return nil
		}

	})

	<-started
	start := time.Now()
	manager.Shutdown()

	duration := time.Since(start)
	if duration > timeout*2 {
		t.Errorf("shutdown took %v, expected <= %v", duration, timeout*2)
	}

}

// TestContextCancellation tests that the context is canceled when the shutdown manager shuts down
func TestContextCancellation(t *testing.T) {

	manager := sm.NewShutdownManager(time.Second)
	serviceCanceled := make(chan struct{})

	manager.Run(func(ctx context.Context) error {
		<-ctx.Done()
		close(serviceCanceled)

		return ctx.Err()
	})

	manager.Shutdown()

	select {
	case <-serviceCanceled:
		// Service was canceled successfully
	case <-time.After(2 * time.Second):
		t.Fatal("service was not canceled")
	}

}
