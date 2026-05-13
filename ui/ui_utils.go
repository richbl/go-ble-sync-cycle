package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// getSessionConfigDir returns the directory path for session configuration files, using
// os.UserConfigDir(), which follows the XDG Base Directory specification
func getSessionConfigDir() (string, error) {

	configHome, err := os.UserConfigDir()
	if err != nil {

		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	configDir := filepath.Join(configHome, ApplicationID)

	// Ensure the directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {

			return "", fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	return configDir, nil
}

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
			logger.Info(logger.BackgroundCtx, logger.GUI, "user confirmed application exit, so...")
			ui.shutdownMgr.Shutdown()
		},
	)
}

// bindValidator ties regex validation check to an EntryRow widget
func bindValidator(entry *adw.EntryRow, pattern string, onUpdate func()) {

	reg := regexp.MustCompile(pattern)

	entry.Connect("changed", func() {

		text := entry.Text()

		// Toggle error class based on regex
		if reg.MatchString(text) {
			entry.RemoveCSSClass("error")
		} else {
			entry.AddCSSClass("error")
		}

		// Trigger the callback if provided (e.g., to update button state)
		if onUpdate != nil {
			onUpdate()
		}

	})

}

// evaluateDisplayTarget checks the requested TOML display name against active Wayland monitors,
// returning a valid string ONLY if the target is an active secondary monitor, otherwise, it
// returns "" (empty string) to preserve windowed scaling behavior
func evaluateDisplayTarget(requestedName string) string {

	requestedName = strings.TrimSpace(requestedName)
	if requestedName == "" {
		return ""
	}

	// Set default display from GDK
	display := gdk.DisplayGetDefault()
	if display == nil {
		return ""
	}

	monitors := display.Monitors()
	count := monitors.NItems()

	// Iterate through all connected display devices
	for i := range count {

		item := monitors.Item(i)
		if item == nil {
			continue
		}

		mon, ok := item.Cast().(*gdk.Monitor)
		if !ok {
			continue
		}

		if mon.Connector() == requestedName {

			// Index 0 is the primary/default display, so we return "" so mpv doesn't force fullscreen
			if i == 0 {
				logger.Debug(logger.BackgroundCtx, logger.GUI, "Target display '%s' is the primary monitor. Preserving default scaling behavior.", requestedName)

				return ""
			}

			// Found a valid, non-default display
			logger.Debug(logger.BackgroundCtx, logger.GUI, "Target '%s' validated as secondary monitor.", requestedName)

			return requestedName
		}
	}

	// No screen name match, so return "" (default display)
	logger.Debug(logger.BackgroundCtx, logger.GUI, "Target display '%s' not found or inactive. Falling back to primary.", requestedName)

	return ""
}
