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

// displayConfirmationDialog shows a Yes/No dialog with customizable appearance for the Yes button
func displayConfirmationDialog(window *adw.ApplicationWindow, title, message string, appearance adw.ResponseAppearance, onYes func()) {

	const (
		yes = "yes"
		no  = "no"
	)

	dialog := adw.NewAlertDialog(title, message)

	// Default/Cancel setup
	dialog.SetCloseResponse(no)
	dialog.SetDefaultResponse(no)

	dialog.AddResponse(no, "No")
	dialog.AddResponse(yes, "Yes")

	// Set specific appearance for the "Yes" action (Suggested or Destructive)
	dialog.SetResponseAppearance(yes, appearance)

	dialog.ConnectResponse(func(response string) {
		if response == yes {
			onYes()
		}
	})

	dialog.Present(gtk.Widgetter(window))

}

// createExitDialog creates the application exit confirmation dialog
func (ui *AppUI) createExitDialog() {
	displayConfirmationDialog(
		ui.Window,
		"Exit BLE Sync Cycle?",
		"Are you sure you want to exit?",
		adw.ResponseDestructive,
		func() {
			logger.Info(logger.BackgroundCtx, logger.GUI, "user confirmed exit")
			ui.shutdownMgr.Shutdown()
		},
	)
}
