package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
)

// setupGUIApplication initializes the GTK UI and sets up all signal handlers
func setupGUIApplication(app *gtk.Application, shutdownMgr *services.ShutdownManager) {

	adw.Init()
	builder := gtk.NewBuilderFromString(uiXML)
	ui := NewAppUI(builder)
	ui.shutdownMgr = shutdownMgr

	// Create the "About" menu item action handler
	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(_ *glib.Variant) {

		// Create a new About dialog
		aboutDialog := adw.NewAboutDialog()
		aboutDialog.SetApplicationIcon("bsc_logo")
		aboutDialog.SetApplicationName("BLE Sync Cycle")
		aboutDialog.SetVersion("v" + config.GetVersion())
		aboutDialog.SetCopyright("Copyright © 2025 Rich Bloch")
		aboutDialog.SetDevelopers([]string{"Rich Bloch"})
		aboutDialog.SetIssueURL("https://github.com/richbl/go-ble-sync-cycle/issues/new?template=issue-report.md")
		aboutDialog.SetLicenseType(gtk.LicenseMITX11)
		aboutDialog.SetWebsite("https://github.com/richbl/go-ble-sync-cycle")

		var transientParent gtk.Widgetter = ui.Window
		aboutDialog.Present(transientParent)

	})

	app.AddAction(aboutAction)

	// Create the "Exit" menu item action handler
	exitAction := gio.NewSimpleAction("exit", nil)
	exitAction.ConnectActivate(func(_ *glib.Variant) {
		logger.Info(logger.BackgroundCtx, logger.GUI, "Exit action triggered from app menu item")
		ui.createExitDialog()
	})

	app.AddAction(exitAction)

	// Set up close request handler (system-level exit via the [x] button)
	ui.Window.ConnectCloseRequest(func() bool {

		safeUpdateUI(func() {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "Exit action triggered from window manager close button")
			ui.createExitDialog()
		})

		return true
	})

	// Create SessionController and initialize (Page 1)
	sessionCtrl := NewSessionController(ui, shutdownMgr)
	sessionCtrl.scanForSessions()
	sessionCtrl.PopulateSessionList()

	setupAllSignals(sessionCtrl)
	ui.Window.SetApplication(app)
	ui.Window.Present()

}
