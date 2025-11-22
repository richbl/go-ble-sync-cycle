package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Session represents the configuration file and its display name
type Session struct {
	ID         int
	Title      string
	ConfigPath string
}

// SessionController manages the logic for Page 1 (Session Selection).
type SessionController struct {
	UI       *AppUI
	Sessions []Session
}

// NewSessionController creates the controller
func NewSessionController(ui *AppUI) *SessionController {
	return &SessionController{
		UI: ui,
	}
}

// --- File Scanning and Data Population Logic ---

// TODO readSessionTitleFromTOML is a STUB function
// REPLACE this with the actual logic from the /internal/config package

func (*SessionController) readSessionTitleFromTOML(filePath string) (string, error) {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	base := filepath.Base(filePath)
	title := strings.TrimSuffix(base, filepath.Ext(base))

	if len(title) > 0 {
		return title, nil
	}

	return "", fmt.Errorf("validation failed: empty session title")
}

// scanForSessions clears the current session list and scans the CWD for valid TOML files
func (sc *SessionController) scanForSessions() {

	sc.Sessions = nil

	// Look for TOML files in the current working directory
	files, err := filepath.Glob("*.toml")
	if err != nil {
		fmt.Printf("Error scanning for TOML files: %v\n", err)
		return
	}

	for i, filePath := range files {
		title, err := sc.readSessionTitleFromTOML(filePath)
		if err != nil {
			fmt.Printf("Skipping invalid TOML file %s: %v\n", filePath, err)
			continue
		}

		session := Session{
			ID:         i + 1,
			Title:      title,
			ConfigPath: filePath,
		}
		sc.Sessions = append(sc.Sessions, session)
	}

}

// PopulateSessionList clears all rows from the ListBox and repopulates it with the current Sessions data
func (sc *SessionController) PopulateSessionList() {

	// Clear existing rows
	sc.UI.Page1.ListBox.RemoveAll()

	// Populate with current sessions
	for _, s := range sc.Sessions {
		row := adw.NewActionRow()
		row.SetTitle(s.Title)
		row.SetSubtitle(fmt.Sprintf("Config: %s", s.ConfigPath))

		sc.UI.Page1.ListBox.Append(row)
	}

	// 3. Ensure buttons are disabled if no row is selected
	sc.UI.Page1.EditButton.SetSensitive(false)
	sc.UI.Page1.LoadButton.SetSensitive(false)
}

// SetupSignals wires up the event listeners for session selection and navigation
func (sc *SessionController) SetupSignals() {

	sc.UI.ViewStack.Connect("notify::visible-child-name", func() {
		if sc.UI.ViewStack.VisibleChildName() == "page1" {
			fmt.Println("View switched to Page 1: Refreshing Session List from CWD...")
			sc.scanForSessions()
			sc.PopulateSessionList()
		}
	})

	// Handle row selection in the ListBox
	sc.UI.Page1.ListBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		hasSelection := (row != nil)

		sc.UI.Page1.EditButton.SetSensitive(hasSelection)
		sc.UI.Page1.LoadButton.SetSensitive(hasSelection)

		if hasSelection {
			idx := row.Index()

			if idx >= 0 && idx < len(sc.Sessions) {
				selectedSession := sc.Sessions[idx]
				fmt.Printf("Selected Session: %s (Config: %s)\n", selectedSession.Title, selectedSession.ConfigPath)
			}
		}
	})

	// Handle load button
	sc.UI.Page1.LoadButton.ConnectClicked(func() {

		selectedRow := sc.UI.Page1.ListBox.SelectedRow()

		if selectedRow != nil {
			idx := selectedRow.Index()

			if idx >= 0 && idx < len(sc.Sessions) {
				selectedSession := sc.Sessions[idx]
				fmt.Printf("Loading Session: %s. Navigating to page 2...\n", selectedSession.Title)

				// TODO: Pass the selectedSession to Page 2 logic here

				sc.UI.ViewStack.SetVisibleChildName("page2")
			}
		}
	})

	// Handle edit button
	sc.UI.Page1.EditButton.ConnectClicked(func() {

		fmt.Println("Navigating to Editor (page 4)...")
		sc.UI.ViewStack.SetVisibleChildName("page4")
	})

}
