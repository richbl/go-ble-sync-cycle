// Package services provides foundational service utilities for the application
//
// Services houses the ShutdownManager, which coordinates the graceful termination
// of the application, ensuring that resources are cleaned up and goroutines exit
// properly upon receiving system signals (e.g., SIGTERM, SIGINT) or internal errors
package services
