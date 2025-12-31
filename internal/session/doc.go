// Package session orchestrates the core workflow of the BSC application
//
// It provides the Manager, which is responsible for:
//   - Loading and validating session configurations
//   - Initializing and synchronizing controllers (BLE, Video, Speed)
//   - Managing the application state machine (Running, Stopped, Editing)
//   - Coordinating the clean shutdown of all active components
//
// The session package acts as the glue that binds the configuration, hardware interfaces,
// and user interface together
package session
