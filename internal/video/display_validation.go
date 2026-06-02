package video

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// ValidateDisplay checks the requested display name against active monitors.
func ValidateDisplay(ctx context.Context, requestedName string) config.DisplayValidationResult {
	requestedName = strings.TrimSpace(requestedName)
	result := config.DisplayValidationResult{
		ActualDisplayName: requestedName,
	}

	if requestedName == "" {
		return result
	}

	// Ensure GTK is initialized (safe to call multiple times, required for CLI mode)
	gtk.Init()

	display := gdk.DisplayGetDefault()
	if display == nil {
		return result
	}

	return findDisplayMonitor(ctx, requestedName, display.Monitors())
}

// findDisplayMonitor iterates over available monitors to find a match
func findDisplayMonitor(ctx context.Context, requestedName string, monitors *gio.ListModel) config.DisplayValidationResult {
	count := monitors.NItems()
	var available []string

	for i := range count {
		item := monitors.Item(i)
		if item == nil {
			continue
		}

		mon, ok := item.Cast().(*gdk.Monitor)
		if !ok {
			continue
		}

		connector := mon.Connector()
		if connector != "" {
			available = append(available, connector)
		}

		if connector == requestedName {
			return processMatchedDisplay(ctx, requestedName, i)
		}
	}

	logger.Debug(ctx, logger.VIDEO, fmt.Sprintf("target display '%s' not found among available monitors %v — falling back to default", requestedName, available))

	return config.DisplayValidationResult{
		IsValid:             false,
		IsNonDefaultMonitor: false,
		ActualDisplayName:   requestedName,
	}
}

// processMatchedDisplay builds the result for a matched display
func processMatchedDisplay(ctx context.Context, requestedName string, index uint) config.DisplayValidationResult {
	result := config.DisplayValidationResult{
		IsValid:           true,
		ActualDisplayName: requestedName,
	}

	// Use name-prefix heuristic: connectors starting with "e" are embedded/default
	result.IsNonDefaultMonitor = !strings.HasPrefix(requestedName, "e")

	if result.IsNonDefaultMonitor {
		logger.Debug(ctx, logger.VIDEO, fmt.Sprintf("target display '%s' validated as non-default monitor (index %d)", requestedName, index))
	} else {
		logger.Debug(ctx, logger.VIDEO, fmt.Sprintf("target display '%s' validated as default/embedded monitor (index %d)", requestedName, index))
	}

	return result
}
