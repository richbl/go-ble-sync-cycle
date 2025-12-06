// File: ui_session_select.go
package ui

import (
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/session"
)

// Session represents the configuration file and its display name
type Session struct {
	ID         int
	Title      string
	ConfigPath string
}

// Status represents the logical connection/battery status
type Status int

const (
	StatusConnected Status = iota
	StatusNotConnected
	StatusStopped
	StatusConnecting
	StatusFailed
)

// ObjectKind represents the logical object we are displaying
type ObjectKind int

const (
	ObjectBLE ObjectKind = iota
	ObjectBattery
)

// StatusPresentation holds the UI-facing data for a status
type StatusPresentation struct {
	Display string
	Icon    string
}

// statusTable centralizes all mappings of (object, status) -> UI data
var statusTable = map[ObjectKind]map[Status]StatusPresentation{
	ObjectBLE: {
		StatusConnected:    {Display: "Connected", Icon: "bluetooth-symbolic"},
		StatusNotConnected: {Display: "Not Connected", Icon: "bluetooth-disconnected-symbolic"},
		StatusStopped:      {Display: "Stopped", Icon: "bluetooth-disconnected-symbolic"},
		StatusConnecting:   {Display: "Connecting...", Icon: "bluetooth-acquiring-symbolic"},
		StatusFailed:       {Display: "Failed", Icon: "bluetooth-disconnected-symbolic"},
	},
	ObjectBattery: {
		StatusConnected:    {Display: "Connected", Icon: "battery-good-symbolic"},
		StatusNotConnected: {Display: "Unknown", Icon: "battery-symbolic"},
		StatusStopped:      {Display: "Unknown", Icon: "battery-symbolic"},
		StatusConnecting:   {Display: "Connecting...", Icon: "battery-symbolic"},
		StatusFailed:       {Display: "Unknown", Icon: "battery-symbolic"},
	},
}

// SessionController manages the logic for Page 1 (Session Selection) and related UI
type SessionController struct {
	UI             *AppUI
	Sessions       []Session
	SessionManager *session.Manager
	starting       atomic.Bool // To prevent multiple concurrent starts
}

// NewSessionController creates the controller
func NewSessionController(ui *AppUI) *SessionController {

	return &SessionController{
		UI:             ui,
		SessionManager: session.NewManager(),
	}
}

// setupSessionSelectSignals wires up event listeners for the session selection tab (Page 1)
func (sc *SessionController) setupSessionSelectSignals() {

	sc.setupListBoxSignals()
	sc.setupLoadButtonSignals()
	sc.setupEditButtonSignals()

}

// scanForSessions looks for valid session config files in the current working directory
func (sc *SessionController) scanForSessions() {

	logger.Debug(logger.GUI, "scanning for session configuration files...")

	sc.Sessions = nil

	// Find all .toml files in the current directory
	files, err := filepath.Glob("*.toml")
	if err != nil {
		logger.Error(logger.GUI, fmt.Sprintf("pattern-matching error when scanning for sessions: %v", err))

		return
	}

	// Check if any files were actually found
	if files == nil || len(files) == 0 {
		logger.Debug(logger.GUI, "no session configuration files found")
		displayAlertDialog(sc.UI.Window, "Session Selection Failed", "No BSC session configuration files found in the current directory")

		return
	}

	// Load metadata for each session file found
	sessionID := 1
	for _, filePath := range files {
		metadata, err := config.LoadSessionMetadata(filePath)

		if err != nil {
			logger.Warn(logger.GUI, fmt.Sprintf("skipping invalid config file %s: %v", filePath, err))
			continue
		}

		if metadata.IsValid {

			session := Session{
				ID:         sessionID,
				Title:      metadata.Title,
				ConfigPath: metadata.FilePath,
			}

			sc.Sessions = append(sc.Sessions, session)
			sessionID++
		}

	}

	logger.Debug(logger.GUI, fmt.Sprintf("session scan complete: found %d valid session(s)", len(sc.Sessions)))

}

// PopulateSessionList refreshes the ListBox with current sessions
func (sc *SessionController) PopulateSessionList() {

	// Clear existing rows
	sc.UI.Page1.ListBox.RemoveAll()

	if sc.Sessions == nil || len(sc.Sessions) == 0 {
		logger.Debug(logger.GUI, "no sessions to populate in the list")

		row := adw.NewActionRow()
		row.SetTitle("No sessions found")
		row.SetSubtitle(fmt.Sprintf(""))
		sc.UI.Page1.ListBox.Append(row)

		return
	}

	logger.Debug(logger.GUI, fmt.Sprintf("populating session list with %d session(s)...", len(sc.Sessions)))

	// Populate with current sessions
	for _, s := range sc.Sessions {
		row := adw.NewActionRow()
		row.SetTitle(s.Title)
		row.SetSubtitle(fmt.Sprintf("config: %s", s.ConfigPath))
		sc.UI.Page1.ListBox.Append(row)
	}

	// Ensure buttons are disabled if no row is selected
	sc.UI.Page1.EditButton.SetSensitive(false)
	sc.UI.Page1.LoadButton.SetSensitive(false)

}

// setupListBoxSignals wires up event listeners for the ListBox
func (sc *SessionController) setupListBoxSignals() {

	sc.UI.Page1.ListBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {

		hasSelection := (row != nil)
		sc.UI.Page1.EditButton.SetSensitive(hasSelection)
		sc.UI.Page1.LoadButton.SetSensitive(hasSelection)

		if hasSelection {
			idx := row.Index()
			if idx >= 0 && idx < len(sc.Sessions) {
				selectedSession := sc.Sessions[idx]
				logger.Debug(logger.GUI, fmt.Sprintf("selected Session: %s (config file: %s)", selectedSession.Title, selectedSession.ConfigPath))
			}

		}

	})

}

// setupLoadButtonSignals wires up event listeners for the Load button
func (sc *SessionController) setupLoadButtonSignals() {

	sc.UI.Page1.LoadButton.ConnectClicked(func() {

		selectedRow := sc.UI.Page1.ListBox.SelectedRow()
		if selectedRow != nil {

			idx := selectedRow.Index()
			if idx >= 0 && idx < len(sc.Sessions) {

				selectedSession := sc.Sessions[idx]
				logger.Debug(logger.GUI, fmt.Sprintf("loading Session: %s...", selectedSession.Title))

				// Load the session into the SessionManager
				err := sc.SessionManager.LoadSession(selectedSession.ConfigPath)
				if err != nil {
					logger.Error(logger.GUI, fmt.Sprintf("error loading session: %v", err))
					return
				}

				logger.Debug(logger.GUI, fmt.Sprintf("session loaded successfully. State: %s", sc.SessionManager.GetState()))

				// Update Page 2 with session info
				sc.updatePage2WithSession(selectedSession)

				// Navigate to Page 2
				sc.UI.ViewStack.SetVisibleChildName("page2")
			}
		}

	})

}

// setupEditButtonSignals wires up event listeners for the Edit button
func (sc *SessionController) setupEditButtonSignals() {

	sc.UI.Page1.EditButton.ConnectClicked(func() {
		logger.Debug(logger.GUI, "navigating to Session Editor (page 4)...")

		// Navigate to Page 4
		sc.UI.ViewStack.SetVisibleChildName("page4")
	})

}
