package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// safeUpdateUI helper for main-thread GUI calls
func safeUpdateUI(fn func()) {

	glib.IdleAdd(func() bool {
		fn()

		return false
	})

}

// setupNavigationSignals sets up ViewStack navigation handlers, with per-page actions
func setupNavigationSignals(stack *adw.ViewStack, pageActions map[string]func()) {

	stack.Connect("notify::visible-child-name", func() {
		pageName := stack.VisibleChildName()
		if action, ok := pageActions[pageName]; ok {
			action()
		}
	})

}

// displayAlertDialog shows a simple alert dialog with an OK button
func displayAlertDialog(window *adw.ApplicationWindow, title, message string) {

	dialog := adw.NewAlertDialog(title, message)
	dialog.AddResponse("ok", "OK")
	dialog.SetResponseAppearance("ok", adw.ResponseSuggested)
	dialog.SetCloseResponse("ok")
	dialog.Present(gtk.Widgetter(window))

}

// createExitDialog creates the application exit confirmation dialog
func (ui *AppUI) createExitDialog() {

	const (
		yes = "yes"
		no  = "no"
	)

	ui.exitDialog = adw.NewAlertDialog(
		"Exit BLE Sync Cycle?",
		"Are you sure you want to exit?",
	)

	// Set dialog properties
	ui.exitDialog.SetCloseResponse(no)
	ui.exitDialog.SetDefaultResponse(no)

	// Add response buttons in the recommended order for GTK4/Adwaita
	ui.exitDialog.AddResponse(no, "No")
	ui.exitDialog.AddResponse(yes, "Yes")

	// Style the Yes button as destructive
	ui.exitDialog.SetResponseAppearance(yes, adw.ResponseDestructive)

	// Connect response signal
	ui.exitDialog.ConnectResponse(func(response string) {

		if response == "yes" {
			logger.Info(logger.BackgroundCtx, logger.GUI, "user confirmed exit")
			// Trigger shutdown via the ShutdownManager to ensure proper cleanup
			ui.shutdownMgr.Shutdown()
		} else {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "user cancelled exit")
		}

	})

	ui.exitDialog.Present(ui.Window)

}
