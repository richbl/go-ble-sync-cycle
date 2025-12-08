package ui

import (
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// setupSessionEditSignals wires up event listeners for the editor tab (Page 4)
func (sc *SessionController) setupSessionEditSignals() {

	// Placeholder: Connect save buttons to serialize UI fields to TOML
	// Load current session config into fields on page enter (e.g., via view stack notify)
	// For now, stub with basic navigation confirmation

	sc.UI.ViewStack.Connect("notify::visible-child-name", func() {

		if sc.UI.ViewStack.VisibleChildName() == "page4" {
			logger.Debug(logger.GUI, "Session Edit: load session config into fields")
			// TODO: Populate sc.UI.Page4 fields from sc.SessionManager.Config()
		}

	})

	// Connect Save button
	sc.UI.Page4.SaveButton.ConnectClicked(func() {
		logger.Debug(logger.GUI, "Session Edit: Save button clicked (stubbed)")
		// TODO: Serialize fields to current config path
	})

	// Connect Save As button
	sc.UI.Page4.SaveAsButton.ConnectClicked(func() {
		logger.Debug(logger.GUI, "Session Edit: Save As button clicked (stubbed)")
		// TODO: Prompt for new path, serialize
	})

	logger.Debug(logger.GUI, "Session Edit: signals setup complete (stubbed)")

}
