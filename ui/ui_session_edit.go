package ui

import (
	"fmt"
)

// setupEditSignals wires up event listeners for the editor tab (Page 4)
func (sc *SessionController) setupEditSignals() {

	// Placeholder: Connect save buttons to serialize UI fields to TOML
	// Load current session config into fields on page enter (e.g., via view stack notify)
	// For now, stub with basic navigation confirmation

	sc.UI.ViewStack.Connect("notify::visible-child-name", func() {

		if sc.UI.ViewStack.VisibleChildName() == "page4" {
			fmt.Println("Entered Editor (page 4): Load session config into fields (stubbed)")
			// TODO: Populate sc.UI.Page4 fields from sc.SessionManager.GetConfig()
		}

	})

	// Connect Save button
	sc.UI.Page4.SaveButton.ConnectClicked(func() {
		fmt.Println("Save button clicked (stubbed)")
		// TODO: Serialize fields to current config path
	})

	// Connect Save As button
	sc.UI.Page4.SaveAsButton.ConnectClicked(func() {
		fmt.Println("Save As button clicked (stubbed)")
		// TODO: Prompt for new path, serialize
	})

	fmt.Println("Edit signals setup complete (stubbed)")

}
