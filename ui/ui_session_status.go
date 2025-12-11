package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/session"
)

// setupSessionControlSignals wires up event listeners for the session control button
func (sc *SessionController) setupSessionControlSignals() {

	sc.UI.Page2.SessionControlBtn.ConnectClicked(func() {
		sc.handleSessionControl()
	})

}

// setupSessionStatusSignals wires up event listeners for the session status tab (Page 2)
func (sc *SessionController) setupSessionStatusSignals() {
	sc.setupSessionControlSignals()
}

// handleSessionControl processes clicks on the session control button
func (sc *SessionController) handleSessionControl() {

	currentState := sc.SessionManager.SessionState()

	logger.Debug(logger.GUI, fmt.Sprintf("button clicked: State=%s", currentState))

	if currentState >= session.StateConnecting || sc.starting.Load() {
		sc.handleStop()
		return
	}

	sc.handleStart()

}

// handleStop processes stopping the session
func (sc *SessionController) handleStop() {

	logger.Debug(logger.GUI, "stop branch entered")

	sc.SessionManager.StopSession()

	logger.Debug(logger.GUI, "stop session returned")

	safeUpdateUI(func() {
		logger.Debug(logger.GUI, "updating UI for stop")
		sc.updateSessionControlButton(false)
		sc.updatePage2Status(StatusStopped, StatusNotConnected, "Unknown")

		// Reset metrics to zero/default
		sc.UI.Page2.SpeedLabel.SetLabel("0.0")
		sc.UI.Page2.PlaybackSpeedLabel.SetLabel("0.00x")
		sc.UI.Page2.TimeRemainingLabel.SetLabel("--:--:--")
	})

}

// handleStart processes starting the session
func (sc *SessionController) handleStart() {

	logger.Info(logger.GUI, "starting BSC Session...")

	if sc.starting.Load() {
		logger.Warn(logger.GUI, "start ignored: already pending")
		return
	}

	if !sc.starting.CompareAndSwap(false, true) {
		logger.Warn(logger.GUI, "start ignored: race on pending")
		return
	}

	// Ensure starting flag is cleared when method exits
	defer sc.starting.Store(false)

	// Update UI to show connecting state
	safeUpdateUI(func() {
		logger.Debug(logger.GUI, "updating UI for start")
		sc.updateSessionControlButton(true)
		sc.updatePage2Status(StatusConnecting, StatusNotConnected, "Unknown")
	})

	// Launch goroutine to start session
	go sc.startSessionGoroutine()

}

// startSessionGoroutine runs the StartSession method and updates UI based on result
func (sc *SessionController) startSessionGoroutine() {

	logger.Debug(logger.GUI, "start goroutine launched")

	defer func() {

		logger.Debug(logger.GUI, "start goroutine exiting")

		safeUpdateUI(func() {
			logger.Debug(logger.GUI, "updating UI post-start")

			// Re-toggle to Start if success/error, but only if stopped
			if sc.SessionManager.SessionState() == session.StateLoaded {
				sc.updateSessionControlButton(false)
			}

		})
	}()

	// Start the session
	logger.Debug(logger.GUI, "calling StartSession()")

	err := sc.SessionManager.StartSession()
	if err != nil {
		sc.handleStartError(err)
		return
	}

	// Update UI to show success state
	logger.Debug(logger.GUI, "StartSession() successful")

	safeUpdateUI(func() {
		logger.Debug(logger.GUI, "updating UI for successful start")
		battery := fmt.Sprintf("%d%%", sc.SessionManager.BatteryLevel())
		sc.updatePage2Status(StatusConnected, StatusConnected, battery)

		sc.startMetricsLoop() // Start metrics loop
	})

}

// handleStartError processes errors from StartSession
func (sc *SessionController) handleStartError(err error) {

	logger.Error(logger.GUI, fmt.Sprintf("start session error: %v", err))

	safeUpdateUI(func() {

		logger.Debug(logger.GUI, "updating UI for error")

		sc.updateSessionControlButton(false)
		if errors.Is(err, context.Canceled) {
			sc.updatePage2Status(StatusStopped, StatusNotConnected, "Unknown")
			return
		}

		// Show error state in UI
		sc.updatePage2Status(StatusFailed, StatusNotConnected, "Unknown")
		displayAlertDialog(sc.UI.Window, "Start Session Failed", err.Error())
	})

}

// updatePage2WithSession refreshes Page 2 UI elements with the given session data
func (sc *SessionController) updatePage2WithSession(sess Session) {

	// Update session name
	sc.UI.Page2.SessionNameRow.SetTitle(sess.Title)

	// Get config from SessionManager
	cfg := sc.SessionManager.Config()
	if cfg == nil {
		return
	}

	// Initial state: BLE not connected, Battery unknown
	sc.updatePage2Status(StatusNotConnected, StatusNotConnected, "Unknown")

	// Reset metrics
	sc.UI.Page2.SpeedLabel.SetLabel("0.0")
	sc.UI.Page2.PlaybackSpeedLabel.SetLabel("0.00x")
	sc.UI.Page2.TimeRemainingLabel.SetLabel("--:--:--")

	// Set button to start mode
	sc.updateSessionControlButton(false)

	// Enable the button now that session is loaded
	sc.UI.Page2.SessionControlBtn.SetSensitive(true)
	logger.Debug(logger.GUI, fmt.Sprintf("page 2 updated with session: %s", sess.Title))

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

}

// setBatteryStatus updates the Battery status indicator on Page 2
func (sc *SessionController) setBatteryStatus(status Status, level string) {

	p := statusTable[ObjectBattery][status]
	display := p.Display

	// If battery is logically connected and a battery level is provided, show the level instead
	if status == StatusConnected && level != "" {
		display = level
	}

	sc.UI.Page2.SensorBatteryRow.SetSubtitle(display)
	sc.UI.Page2.SensorBattIcon.SetFromIconName(p.Icon)

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

		// If session isn't running, stop the loop (return false)
		if sc.SessionManager.SessionState() != session.StateRunning {
			return false
		}

		// Get Data from Manager
		speed, _ := sc.SessionManager.CurrentSpeed()
		timeRem := sc.SessionManager.VideoTimeRemaining()
		rate := sc.SessionManager.VideoPlaybackRate()

		// Update widget labels
		sc.UI.Page2.SpeedLabel.SetLabel(fmt.Sprintf("%.1f", speed))
		sc.UI.Page2.PlaybackSpeedLabel.SetLabel(fmt.Sprintf("%.2fx", rate))
		sc.UI.Page2.TimeRemainingLabel.SetLabel(timeRem)

		// Return true to keep the loop running
		return true
	})
}
