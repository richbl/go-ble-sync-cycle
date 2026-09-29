// Package ui implements the Graphical User Interface (GUI) for the application
//
// Built with GTK4 and Adwaita, it provides a responsive interface organized into
// distinct pages:
//
//   - Session Selection (Page 1):
//     Lists available sessions and provides entry points for creating new ones
//
//   - Session Status (Page 2):
//     The main dashboard during a workout, displaying real-time metrics (speed, cadence,
//     heart rate), video playback status, and connection health
//
//   - Session Log (Page 3):
//     A live console view of application logs, useful for debugging and monitoring
//     sensor events
//
//   - Session Editor (Page 4):
//     A comprehensive form for configuring all aspects of a session, including
//     video paths, sensor addresses, and simulation settings
//
// The UI interacts with the core logic primarily through the session package
package ui
