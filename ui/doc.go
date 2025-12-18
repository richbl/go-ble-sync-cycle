// Package ui provides the complete GTK4 and libadwaita-based graphical user
// interface for the BLE Sync Cycle (BSC) application. It is responsible for
// window creation, widget management, user interaction, and displaying data
// from the application's core logic.

// The UI is structured around a central `AppUI` controller which manages several
// distinct pages:

//   - Session Selection (Page 1): Allows users to view, select, and load
//     existing session configuration files (`.toml`). It also provides entry
//     points for creating or editing sessions

//   - Session Status (Page 2): Displays real-time metrics for an active session,
//     including BLE sensor connection status, battery level, cycling speed, and
//     video playback information. It contains the controls to start and stop a session

//   - Session Log (Page 3): Provides a live view of application logs directly
//     within the UI. It features a custom `io.Writer` that translates colored
//     ANSI log output into styled text in a GtkTextView

//   - Session Editor (Page 4): A comprehensive form for creating a new session
//     or modifying an existing one. It allows configuration of application, BLE,
//     speed, video, and OSD settings, and handles saving the configuration
package ui
