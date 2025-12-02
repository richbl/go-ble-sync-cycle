package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
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
