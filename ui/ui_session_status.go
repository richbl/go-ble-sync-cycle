package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/richbl/go-ble-sync-cycle/internal/ble"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/session"
	"github.com/richbl/go-ble-sync-cycle/internal/video"
)

const (
	errFormat     = "%v: %w"
	StatusUnknown = "unknown"
)

// setupSessionStatusSignals wires up event listeners for the session status tab (Page 2)
func (sc *SessionController) setupSessionStatusSignals() {
	sc.setupSessionControlSignals()
}

// setupSessionControlSignals wires up event listeners for the session control button
func (sc *SessionController) setupSessionControlSignals() {

	sc.UI.Page2.SessionControlBtn.ConnectClicked(func() {

		if err := sc.handleSessionControl(); err != nil {
			logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("failed to handle session control: %v", err))
		}

	})

}

// handleSessionControl processes clicks on the session control button
func (sc *SessionController) handleSessionControl() error {

	// Debounce session control button
	safeUpdateUI(func() {
		sc.UI.Page2.SessionControlBtn.SetSensitive(false)
	})

	safeUpdateUI(func() {
		sc.UI.Page2.SessionControlBtn.SetSensitive(true)
	})

	currentState := sc.SessionManager.SessionState()
	logger.Debug(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("button clicked: State=%s", currentState))

	if currentState >= session.StateConnecting || sc.starting.Load() {

		// Stop the session!
		if err := sc.handleStop(); err != nil {
			return fmt.Errorf(errFormat, "unable to stop session", err)
		}

		return nil
	}

	sc.handleStart()

	return nil
}

// handleStart processes starting the session
func (sc *SessionController) handleStart() {

	logger.Info(logger.BackgroundCtx, logger.GUI, "starting BSC Session...")

	if !sc.starting.CompareAndSwap(false, true) {
		logger.Warn(logger.BackgroundCtx, logger.GUI, "start request ignored: already pending")

		return
	}

	// Update UI to show connecting state
	logger.Debug(logger.BackgroundCtx, logger.GUI, "updating UI for start")

	safeUpdateUI(func() {
		sc.updateSessionControlButton(true)
		sc.updatePage2Status(StatusConnecting, StatusNotConnected, StatusUnknown)
	})

	// Launch goroutine to start session
	go func() {
		defer sc.starting.Store(false)
		sc.startSessionGUI()
	}()

}

// handleStartError processes errors from StartSession
func (sc *SessionController) handleStartError(err error) {

	logger.Debug(logger.BackgroundCtx, logger.GUI, "updating UI for error")

	safeUpdateUI(func() {

		sc.updateSessionControlButton(false)
		if errors.Is(err, context.Canceled) {
			sc.updatePage2Status(StatusStopped, StatusNotConnected, StatusUnknown)

			return
		}

		// Show error state in UI
		sc.updatePage2Status(StatusFailed, StatusNotConnected, StatusUnknown)

		logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("session start failed: %v", err))

		// Check for scanning timeout special case, and show appropriate message
		if errors.Is(err, ble.ErrScanTimeout) {
			sessionConnectTimeout := sc.SessionManager.ActiveConfig().BLE.ScanTimeoutSecs
			displayAlertDialog(sc.UI.Window, "BSC Session Start Timeout", fmt.Sprintf("Failed to start the BSC Session due to a BLE device timeout (%ds).\n\nPlease restart the BSC Session.", sessionConnectTimeout))
		} else {
			displayAlertDialog(sc.UI.Window, "Start Session Failed", "Failed to start the BSC Session.\n\nPlease review the BSC Session Log for details.")
		}

	})

}

// handleStop processes stopping the session
func (sc *SessionController) handleStop() error {

	logger.Debug(logger.BackgroundCtx, logger.GUI, "stop branch entered")

	if err := sc.SessionManager.StopSession(); err != nil {
		return fmt.Errorf(errFormat, "unable to stop services", err)
	}

	logger.Debug(logger.BackgroundCtx, logger.GUI, "stop session returned")
	logger.Debug(logger.BackgroundCtx, logger.GUI, "updating UI for stop")

	safeUpdateUI(func() {
		sc.updateSessionControlButton(false)
		sc.updatePage2Status(StatusStopped, StatusNotConnected, StatusUnknown)
		sc.resetMetrics()

		// User edited the running session! (so update the details)
		if cfg := sc.SessionManager.ActiveConfig(); cfg != nil {
			sc.UI.Page2.SessionNameRow.SetSubtitle(cfg.App.SessionTitle)
		}

		path := sc.SessionManager.LoadedConfigPath()
		sc.UI.Page2.SessionFileLocationRow.SetSubtitle(path)
	})

	return nil
}

// startSessionGUI runs the StartSession method and updates UI based on result
func (sc *SessionController) startSessionGUI() {

	logger.Debug(logger.BackgroundCtx, logger.GUI, "start goroutine launched")

	defer func() {
		logger.Debug(logger.BackgroundCtx, logger.GUI, "start goroutine exiting")
		logger.Debug(logger.BackgroundCtx, logger.GUI, "updating UI post-start")

		safeUpdateUI(func() {
			// Re-toggle to Start if success/error, but only if stopped
			if sc.SessionManager.SessionState() == session.StateLoaded {
				sc.updateSessionControlButton(false)
			}

		})

	}()

	// Start the session
	logger.Debug(logger.BackgroundCtx, logger.GUI, "calling StartSession()")
	err := sc.SessionManager.StartSession()
	if err != nil {
		sc.handleStartError(err)

		return
	}

	// Update UI to show success state
	logger.Debug(logger.BackgroundCtx, logger.GUI, "StartSession() successful")
	logger.Debug(logger.BackgroundCtx, logger.GUI, "updating UI for successful start")

	safeUpdateUI(func() {
		battery := fmt.Sprintf("%d%%", sc.SessionManager.BatteryLevel())
		sc.updatePage2Status(StatusConnected, StatusConnected, battery)
		sc.startMetricsLoop()
	})

}

// updatePage2WithSession refreshes Page 2 UI elements with the given session data
func (sc *SessionController) updatePage2WithSession(sess Session) {

	// Update session name and file location
	sc.UI.Page2.SessionNameRow.SetSubtitle(sess.Title)
	sc.UI.Page2.SessionFileLocationRow.SetSensitive(true)
	sc.UI.Page2.SessionFileLocationRow.SetSubtitle(sess.ConfigPath)
	sc.UI.Page2.SessionNameRow.SetSensitive(true)

	// Initial state: BLE not connected, Battery unknown
	sc.updatePage2Status(StatusNotConnected, StatusNotConnected, StatusUnknown)
	sc.resetMetrics()

	// Enable BLE section controls
	sc.UI.Page2.SensorStatusRow.SetSensitive(true)
	sc.UI.Page2.SensorBatteryRow.SetSensitive(true)

	// Enable session metrics controls
	sc.UI.Page2.SpeedRow.SetSensitive(true)
	sc.UI.Page2.PlaybackSpeedRow.SetSensitive(true)
	sc.UI.Page2.TimeRemainingRow.SetSensitive(true)

	// Set button to start mode
	sc.updateSessionControlButton(false)

	// Enable the button now that session is loaded
	sc.UI.Page2.SessionControlRow.SetSensitive(true)

	logger.Debug(logger.BackgroundCtx, logger.GUI, "page 2 updated with session: "+sess.Title)

}

// resetMetrics resets the metrics on Page 2
func (sc *SessionController) resetMetrics() {

	sc.UI.Page2.SpeedLabel.SetLabel("0.0")
	sc.UI.Page2.PlaybackSpeedLabel.SetLabel("0.00x")
	sc.UI.Page2.TimeRemainingLabel.SetLabel("--:--:--")

}

// updatePage2Status updates the BLE and Battery status indicators on Page 2
func (sc *SessionController) updatePage2Status(bleStatus Status, batteryStatus Status, batteryLevel string) {

	sc.setBLEStatus(bleStatus)
	sc.setBatteryStatus(batteryStatus, batteryLevel)

}

// setBLEStatus updates the BLE status indicator on Page 2
func (sc *SessionController) setBLEStatus(status Status) {

	p := statusTable[ObjectBLE][status]
	sc.UI.Page2.SensorStatusRow.SetSubtitle(p.Display)
	sc.UI.Page2.SensorConnIcon.SetFromIconName(p.Icon)
	sc.UI.Page2.SensorConnIcon.SetCSSClasses([]string{p.CSSStyle})

}

// setBatteryStatus updates the Battery status indicator on Page 2
func (sc *SessionController) setBatteryStatus(status Status, level string) {

	p := statusTable[ObjectBattery][status]
	display := p.Display

	// If battery is logically connected and a battery level is provided, show the level
	if status == StatusConnected && level != "" {
		display = level
	}

	sc.UI.Page2.SensorBatteryRow.SetSubtitle(display)
	sc.UI.Page2.SensorBattIcon.SetFromIconName(p.Icon)
	sc.UI.Page2.SensorBattIcon.SetCSSClasses([]string{p.CSSStyle})

}

// updateSessionControlButton updates the session control button label and icon
func (sc *SessionController) updateSessionControlButton(isRunning bool) {

	if isRunning {
		sc.UI.Page2.SessionControlBtnContent.SetLabel("Stop Session")
		sc.UI.Page2.SessionControlBtnContent.SetIconName("media-playback-stop-symbolic")
	} else {
		sc.UI.Page2.SessionControlBtnContent.SetLabel("Start Session")
		sc.UI.Page2.SessionControlBtnContent.SetIconName("media-playback-start-symbolic")
	}

}

// startMetricsLoop initiates a GLib timeout to poll the SessionManager for real-time data
func (sc *SessionController) startMetricsLoop() {

	// Poll every 250ms
	sc.metricsLoop = glib.TimeoutAdd(250, func() bool {

		state := sc.SessionManager.SessionState()

		// Check for async failure (e.g., invalid video file)
		if state == session.StateError {

			logger.Debug(logger.BackgroundCtx, logger.GUI, "metrics loop detected session error")

			// Differentiate between a crash (invalid file) and EOF (video complete) based on error message
			if strings.Contains(sc.SessionManager.ErrorMessage(), video.ErrVideoComplete.Error()) {
				displayAlertDialog(sc.UI.Window, "The BSC Session has Ended", "The video playback has finished.\n\nSession stopped.")
			} else {
				displayAlertDialog(sc.UI.Window, "BSC Session Load Failed", "Invalid video file format. Edit the session file and try again.\n\nPlease review the BSC Session Log for details.")
			}

			// Reset UI and application state
			if err := sc.handleStop(); err != nil {
				logger.Error(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("failed to clean up after session error: %v", err))
			}

			return false
		}

		// If session isn't running, stop the loop
		if state != session.StateRunning {
			return false
		}

		// Update metrics
		speed, _ := sc.SessionManager.CurrentSpeed()
		timeRem := sc.SessionManager.VideoTimeRemaining()
		rate := sc.SessionManager.VideoPlaybackRate()

		// Update widget labels
		sc.UI.Page2.SpeedLabel.SetLabel(fmt.Sprintf("%.1f", speed))
		sc.UI.Page2.PlaybackSpeedLabel.SetLabel(fmt.Sprintf("%.2fx", rate))
		sc.UI.Page2.TimeRemainingLabel.SetLabel(timeRem)

		// Return true to keep the loop chugging along...
		return true
	})

}
