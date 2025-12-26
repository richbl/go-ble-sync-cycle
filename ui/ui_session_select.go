// File: ui_session_select.go
package ui

import (
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
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

const (
	// BLE icons
	iconBLEConnected    = "bluetooth-symbolic"
	iconBLENotConnected = "bluetooth-disconnected-symbolic"
	iconBLEConnecting   = "bluetooth-acquiring-symbolic"

	// Battery icons
	iconBatteryConnected    = "battery-good-symbolic"
	iconBatteryNotConnected = "battery-symbolic"
	iconBatteryConnecting   = "battery-symbolic"
)

// statusTable centralizes all mappings of (object, status) -> UI data
var statusTable = map[ObjectKind]map[Status]StatusPresentation{
	ObjectBLE: {
		StatusConnected:    {Display: "Connected", Icon: iconBLEConnected},
		StatusNotConnected: {Display: "Not Connected", Icon: iconBLENotConnected},
		StatusStopped:      {Display: "Stopped", Icon: iconBLENotConnected},
		StatusConnecting:   {Display: "Connecting...", Icon: iconBLEConnecting},
		StatusFailed:       {Display: "Failed", Icon: iconBLENotConnected},
	},
	ObjectBattery: {
		StatusConnected:    {Display: "Connected", Icon: iconBatteryConnected},
		StatusNotConnected: {Display: "Unknown", Icon: iconBatteryNotConnected},
		StatusStopped:      {Display: "Unknown", Icon: iconBatteryNotConnected},
		StatusConnecting:   {Display: "Connecting...", Icon: iconBatteryConnecting},
		StatusFailed:       {Display: "Unknown", Icon: iconBatteryNotConnected},
	},
}

// SessionController manages the logic for Page 1 (Session Selection) and related UI
type SessionController struct {
	UI             *AppUI
	Sessions       []Session
	SessionManager *session.StateManager
	starting       atomic.Bool
	metricsLoop    glib.SourceHandle
	saveFileDialog *gtk.FileDialog
}

// NewSessionController creates the controller
func NewSessionController(ui *AppUI) *SessionController {

	return &SessionController{
		UI:             ui,
		SessionManager: session.NewManager(),
	}
}

// PopulateSessionList refreshes the ListBox with current sessions
func (sc *SessionController) PopulateSessionList() {

	// Clear existing rows
	sc.UI.Page1.ListBox.RemoveAll()

	if len(sc.Sessions) == 0 {
		logger.Debug(logger.BackgroundCtx, logger.GUI, "no sessions to populate in the list")

		row := adw.NewActionRow()
		row.SetTitle("No sessions found")
		row.SetSubtitle("")
		sc.UI.Page1.ListBox.Append(row)

		return
	}

	// Enable the list of sessions
	sc.UI.Page1.ListBox.SetSensitive(true)

	logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("populating session list with %d session(s)...", len(sc.Sessions)))

	// Populate with current sessions
	for _, s := range sc.Sessions {
		row := adw.NewActionRow()
		row.SetTitle(s.Title)
		sc.UI.Page1.ListBox.Append(row)
	}

	// Set focus on first element
	sc.UI.Page1.ListBox.SelectRow(sc.UI.Page1.ListBox.RowAtIndex(0))

	// With session selection made, enable buttons
	sc.UI.Page1.EditButton.SetSensitive(true)
	sc.UI.Page1.LoadButton.SetSensitive(true)

}

// setupSessionSelectSignals wires up event listeners for the session selection tab (Page 1)
func (sc *SessionController) setupSessionSelectSignals() {

	sc.setupListBoxSignals()
	sc.setupLoadButtonSignals()
	sc.setupEditButtonSignals()

}

// scanForSessions looks for valid session config files in the current working directory
func (sc *SessionController) scanForSessions() {

	logger.Debug(logger.BackgroundCtx, logger.GUI, "scanning for session configuration files...")

	sc.Sessions = nil

	// Find all .toml files in the current directory
	files, err := filepath.Glob("*.toml")
	if err != nil {
		logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("pattern-matching error when scanning for sessions: %v", err))

		return
	}

	// Check if any files were actually found
	if len(files) == 0 {
		logger.Info(logger.BackgroundCtx, logger.GUI, "no session configuration files found")
		displayAlertDialog(sc.UI.Window, "No BSC Sessions", "No BSC session configuration files found in the current directory")

		return
	}

	// Load metadata for each session file found
	sessionID := 1
	for _, filePath := range files {
		metadata, err := config.LoadSessionMetadata(filePath)

		if err != nil {
			logger.Warn(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("skipping invalid config file %s: %v", filePath, err))

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

	logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("session scan complete: found %d valid session(s)", len(sc.Sessions)))

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
				logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("selected session: %s (config file: %s)", selectedSession.Title, selectedSession.ConfigPath))
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
				logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("loading Session: %s...", selectedSession.Title))

				// Load the session into the SessionManager
				err := sc.SessionManager.LoadSession(selectedSession.ConfigPath)
				if err != nil {
					logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("error loading session: %v", err))

					return
				}

				logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("session loaded successfully. State: %s", sc.SessionManager.SessionState()))

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
		logger.Debug(logger.BackgroundCtx, logger.GUI, "navigating to Session Editor (page 4)...")

		// Navigate to Page 4
		sc.UI.ViewStack.SetVisibleChildName("page4")
	})

}
