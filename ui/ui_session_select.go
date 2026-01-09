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
	"github.com/richbl/go-ble-sync-cycle/internal/services"
	"github.com/richbl/go-ble-sync-cycle/internal/session"
)

// SessionController manages the logic for Page 1 (Session Selection) and related UI
type SessionController struct {
	UI             *AppUI
	Sessions       []Session
	SessionManager *session.StateManager
	shutdownMgr    *services.ShutdownManager
	starting       atomic.Bool
	metricsLoop    glib.SourceHandle
	saveFileDialog *gtk.FileDialog
}

// NewSessionController creates the controller
func NewSessionController(ui *AppUI, shutdownMgr *services.ShutdownManager) *SessionController {

	return &SessionController{
		UI:             ui,
		SessionManager: session.NewManager(),
		shutdownMgr:    shutdownMgr,
	}
}

// PopulateSessionList refreshes the ListBox with current sessions
func (sc *SessionController) PopulateSessionList() {

	// Clear existing rows (reset list)
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

// scanForSessions looks for valid session config files in the application's config directory
func (sc *SessionController) scanForSessions() {

	logger.Debug(logger.BackgroundCtx, logger.GUI, "scanning for session configuration files...")

	sc.Sessions = nil

	// Get session configuration directory
	configDir, err := getSessionConfigDir()
	if err != nil {
		logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("failed to get session config directory: %v", err))

		return
	}

	// Find all .toml files in the config directory
	files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
	if err != nil {
		logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("pattern-matching error when scanning for sessions: %v", err))

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

	// Check if any files were actually found
	if len(sc.Sessions) == 0 {
		logger.Info(logger.BackgroundCtx, logger.GUI, "no session configuration files found")

		safeUpdateUI(func() {
			displayAlertDialog(sc.UI.Window, "No BSC Sessions", "No configuration files found in the configuration directory")
		})
	}

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
		if selectedRow == nil {
			return
		}

		idx := selectedRow.Index()
		if idx < 0 || idx >= len(sc.Sessions) {
			return
		}
		selectedSession := sc.Sessions[idx]

		// Check if a session is currently running
		if sc.SessionManager.IsRunning() {

			activeTitle := "Unknown"
			if cfg := sc.SessionManager.ActiveConfig(); cfg != nil {
				activeTitle = cfg.App.SessionTitle
			}

			// Show session stop/replace confirmation dialog
			displayConfirmationDialog(
				sc.UI.Window,
				"Stop Current BSC Session?",
				fmt.Sprintf("'%s' is currently running\n\nDo you want to stop and switch to '%s'?", activeTitle, selectedSession.Title),
				adw.ResponseDestructive,
				func() {

					// User confirmed stop
					if err := sc.SessionManager.StopSession(); err != nil {
						logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("failed to stop session: %v", err))

						return
					}
					// Proceed with load
					sc.performLoadSession(selectedSession)
				},
			)

			return
		}
		// Not running, proceed normally
		sc.performLoadSession(selectedSession)
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

// performLoadSession handles the actual loading and navigation logic
func (sc *SessionController) performLoadSession(selectedSession Session) {

	logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("loading Session: %s...", selectedSession.Title))

	// Load the session into the SessionManager
	err := sc.SessionManager.LoadTargetSession(selectedSession.ConfigPath)
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
