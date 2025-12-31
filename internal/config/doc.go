// Package config defines the configuration structures and defaults for the application
//
// It handles loading Settings from TOML files and validating them against application requirements
// The configuration is split into logical sections:
//   - AppConfig: general application settings
//   - SessionConfig: specific settings for the current workout session
//   - BLEConfig: Bluetooth sensor parameters
//   - VideoConfig: video playback specific settings
//   - SpeedConfig: wheel size and simulation settings
//   - OSDConfig: On-Screen Display (OSD) customization
package config
