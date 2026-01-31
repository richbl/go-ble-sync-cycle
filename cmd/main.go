package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/richbl/go-ble-sync-cycle/internal/flags"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
	"github.com/richbl/go-ble-sync-cycle/internal/session"
	"github.com/richbl/go-ble-sync-cycle/ui"
)

// Application constants
const (
	configFile = "config.toml"
)

func main() {

	// Initialize the application
	appInitialize()

	// Hello computer...
	services.WaveHello(logger.BackgroundCtx)

	// Parse for command-line flags
	parseCLIFlags()
	checkForHelpFlag()

	// Check for application mode (CLI or GUI)
	if !flags.IsCLIMode() {
		logger.Debug(logger.BackgroundCtx, logger.APP, "now running in GUI mode...")
		ui.StartGUI()

		return
	}

	// Continue running in CLI mode
	logger.Debug(logger.BackgroundCtx, logger.APP, "running in CLI mode")

	// Create session manager
	sessionMgr := session.NewManager()

	// Load configuration
	if err := sessionMgr.LoadTargetSession(configFile); err != nil {
		logger.Fatal(logger.BackgroundCtx, logger.APP, err)
	}

	// Start the session (initializes controllers, connects BLE, starts services)
	if err := sessionMgr.StartSession(); err != nil {

		if errors.Is(err, context.Canceled) {
			logger.Info(logger.BackgroundCtx, logger.APP, "application exiting due to user cancellation")
		} else {
			logger.Fatal(logger.BackgroundCtx, logger.APP, err)
		}
	}

	// Wait patiently for shutdown (Ctrl+C or services error)
	sessionMgr.Wait()

	// Wave goodbye
	services.WaveGoodbye(logger.BackgroundCtx)

}

// appInitialize defaults the logger and exit handler objects until later services start
func appInitialize() {

	// Initialize the default logger until user-specified config file is loaded
	logger.Initialize("debug")

	// Initialize the fatal log events exit handler until the service manager is loaded
	logger.SetExitHandler(func() {
		services.WaveGoodbye(logger.BackgroundCtx)
	})

}

// parseCLIFlags parses and validates command-line flags
func parseCLIFlags() {

	if err := flags.ParseArgs(); err != nil {
		logger.Fatal(logger.BackgroundCtx, logger.APP, fmt.Sprintf("failed to parse command-line flags: %v", err))
	}

}

// checkForHelpFlag checks for the help flag passed on the command-line
func checkForHelpFlag() {

	if flags.IsHelpFlag() {
		flags.ShowHelp()
		services.WaveGoodbye(logger.BackgroundCtx)
	}

}
