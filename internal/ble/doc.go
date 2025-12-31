// Package ble implements the Bluetooth Low Energy (BLE) interface for BLE Sync Cycle
//
// It provides a central controller that manages the lifecycle of BLE peripheral interactions,
// including:
//   - Scanning for specific BLE devices (filtered by BD_ADDR)
//   - Establishing and maintaining connections to the BLE devices
//   - Binding to required services: Cycling Speed and Cadence (CSC) and Battery Service
//   - Handling notifications for real-time data updates
//
// The package abstracts the underlying BLE implementation (using tinygo.org/x/bluetooth)
// to provide a clean API for the rest of the application
package ble
