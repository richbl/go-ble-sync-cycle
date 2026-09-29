// Package main serves as the entry point for the BLE Sync Cycle (BSC) application
//
// BLE Sync Cycle (BSC) is an application designed to synchronize video playback with real-time
// cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC)
// sensors
//
// This package is responsible for:
//   - Initializing the application's core services and logging
//   - Parsing command-line flags to determine the operating mode (CLI vs GUI)
//   - Bootstrapping the Session Manager for CLI mode or launching the GTK4 user interface
//
// While main handles the initial startup, the actual business logic is delegated to the
// internal/session package, and the graphical interface is managed by the ui package
package main
