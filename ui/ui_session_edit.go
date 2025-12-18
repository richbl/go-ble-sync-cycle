package ui

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// Maps for dropdown list widgets
var (
	logLevels    = []string{"debug", "info", "warn", "error"}
	speedUnits   = []string{"mph", "km/h"}
	mediaPlayers = []string{"mpv", "vlc"}
)

// setupSessionEditSignals wires up event listeners for the Edit button
func (sc *SessionController) setupSessionEditSignals() {

	sc.UI.Page1.EditButton.ConnectClicked(func() {

		// Identify the chosen BSC session
		selectedRow := sc.UI.Page1.ListBox.SelectedRow()
		if selectedRow == nil {
			return
		}

		idx := selectedRow.Index()
		if idx < 0 || idx >= len(sc.Sessions) {
			return
		}

		selectedSession := sc.Sessions[idx]
		logger.Debug(logger.GUI, fmt.Sprintf("preparing to edit Session: %s...", selectedSession.Title))

		// Load the session into the SessionManager
		if err := sc.SessionManager.LoadSession(selectedSession.ConfigPath); err != nil {
			logger.Warn(logger.GUI, fmt.Sprintf("loading session for edit generated warnings: %v", err))
		}

		// Populate the UI with the session data
		logger.Debug(logger.GUI, "populating UI with session data...")
		sc.populateEditor()

		// Connect file chooser for video button
		sc.UI.Page4.VideoFileButton.ConnectClicked(func() {
			logger.Debug(logger.GUI, "Video file button clicked")
			sc.openVideoFilePicker()
		})

		// Connect Save button
		sc.UI.Page4.SaveButton.ConnectClicked(func() {
			sc.saveSession(false) // Save to current path
		})

		// Connect Save As button
		sc.UI.Page4.SaveAsButton.ConnectClicked(func() {
			sc.saveSession(true) // Save As new path
		})

		// Navigate to the Editor (Page 4)
		logger.Debug(logger.GUI, "navigating to Session Editor (page 4)...")
		sc.UI.ViewStack.SetVisibleChildName("page4")

	})

}

// populateEditor fills the UI widgets with data from the current session config
func (sc *SessionController) populateEditor() {

	cfg := sc.SessionManager.Config()
	if cfg == nil {
		logger.Warn(logger.GUI, "attempted to populate editor with nil config")
		return
	}

	logger.Debug(logger.GUI, "populating editor with session data and enabling widgets")

	p4 := sc.UI.Page4

	// --- App Section ---
	p4.TitleEntry.SetText(cfg.App.SessionTitle)
	p4.LogLevel.SetSelected(indexOf(cfg.App.LogLevel, logLevels))

	// --- BLE Section ---
	p4.BTAddressEntry.SetText(cfg.BLE.SensorBDAddr)
	p4.ScanTimeout.SetValue(float64(cfg.BLE.ScanTimeoutSecs))

	// --- Speed Section ---
	p4.WheelCircumference.SetValue(float64(cfg.Speed.WheelCircumferenceMM))
	p4.SpeedUnits.SetSelected(indexOf(cfg.Speed.SpeedUnits, speedUnits))
	p4.SpeedThreshold.SetValue(cfg.Speed.SpeedThreshold)
	p4.SpeedSmoothing.SetValue(float64(cfg.Speed.SmoothingWindow))

	// --- Video Section ---
	p4.MediaPlayer.SetSelected(indexOf(cfg.Video.MediaPlayer, mediaPlayers))
	p4.VideoFileRow.SetSubtitle(cfg.Video.FilePath)
	p4.StartTimeEntry.SetText(cfg.Video.SeekToPosition)
	p4.WindowScale.SetValue(cfg.Video.WindowScaleFactor)
	p4.UpdateInterval.SetValue(cfg.Video.UpdateIntervalSec)
	p4.SpeedMultiplier.SetValue(cfg.Video.SpeedMultiplier)

	// --- OSD Section ---
	p4.SwitchCycleSpeed.SetActive(cfg.Video.OnScreenDisplay.DisplayCycleSpeed)
	p4.SwitchPlaybackSpeed.SetActive(cfg.Video.OnScreenDisplay.DisplayPlaybackSpeed)
	p4.SwitchTimeRemaining.SetActive(cfg.Video.OnScreenDisplay.DisplayTimeRemaining)
	p4.FontSize.SetValue(float64(cfg.Video.OnScreenDisplay.FontSize))
	p4.MarginLeft.SetValue(float64(cfg.Video.OnScreenDisplay.MarginX))
	p4.MarginTop.SetValue(float64(cfg.Video.OnScreenDisplay.MarginY))

	// Enable all widgets
	toggleSensitive(p4, true)
}

// toggleSensitive enables or disables widgets
func toggleSensitive(p4 *PageSessionEditor, enabled bool) {

	// Use reflection to iterate through the widgets and set their sensitivity
	v := reflect.ValueOf(p4).Elem()

	for i := 0; i < v.NumField(); i++ {

		field := v.Field(i)
		if field.CanInterface() {
			widget, ok := field.Interface().(interface{ SetSensitive(bool) })

			if ok {
				widget.SetSensitive(enabled)
			}

		}

	}

}

// harvestEditor gathers data from widgets into a new Config struct
func (sc *SessionController) harvestEditor() *config.Config {

	p4 := sc.UI.Page4
	cfg := &config.Config{}

	// App
	cfg.App.SessionTitle = p4.TitleEntry.Text()
	cfg.App.LogLevel = logLevels[p4.LogLevel.Selected()]

	// BLE
	cfg.BLE.SensorBDAddr = p4.BTAddressEntry.Text()
	cfg.BLE.ScanTimeoutSecs = int(p4.ScanTimeout.Value())

	// Speed
	cfg.Speed.WheelCircumferenceMM = int(p4.WheelCircumference.Value())
	cfg.Speed.SpeedUnits = speedUnits[p4.SpeedUnits.Selected()]
	cfg.Speed.SpeedThreshold = p4.SpeedThreshold.Value()
	cfg.Speed.SmoothingWindow = int(p4.SpeedSmoothing.Value())

	// Video
	cfg.Video.MediaPlayer = mediaPlayers[p4.MediaPlayer.Selected()]
	cfg.Video.FilePath = p4.VideoFileRow.Subtitle()
	cfg.Video.SeekToPosition = p4.StartTimeEntry.Text()
	cfg.Video.WindowScaleFactor = p4.WindowScale.Value()
	cfg.Video.UpdateIntervalSec = p4.UpdateInterval.Value()
	cfg.Video.SpeedMultiplier = p4.SpeedMultiplier.Value()

	// OSD
	cfg.Video.OnScreenDisplay.DisplayCycleSpeed = p4.SwitchCycleSpeed.Active()
	cfg.Video.OnScreenDisplay.DisplayPlaybackSpeed = p4.SwitchPlaybackSpeed.Active()
	cfg.Video.OnScreenDisplay.DisplayTimeRemaining = p4.SwitchTimeRemaining.Active()
	cfg.Video.OnScreenDisplay.FontSize = int(p4.FontSize.Value())
	cfg.Video.OnScreenDisplay.MarginX = int(p4.MarginLeft.Value())
	cfg.Video.OnScreenDisplay.MarginY = int(p4.MarginTop.Value())

	return cfg
}

// openVideoFilePicker opens a native file dialog to select a video
func (sc *SessionController) openVideoFilePicker() {

	logger.Debug(logger.GUI, "Opening video file dialog...")

	fileDialog := gtk.NewFileDialog()
	fileDialog.SetTitle("Select Video File")

	// Set filters
	filter := gtk.NewFileFilter()
	filter.SetName("Video Files")
	filter.AddPattern("*.mp4")
	filter.AddPattern("*.mkv")
	filter.AddPattern("*.avi")

	filters := gio.NewListStore(filter.Type())
	filters.Append(filter.Object)
	fileDialog.SetFilters(filters)

	// Define callback to handle file selection
	cb := func(res gio.AsyncResulter) {
		file, err := fileDialog.OpenFinish(res)
		if err != nil {
			logger.Warn(logger.GUI, fmt.Sprintf("File dialog cancelled or error: %v", err))
			return
		}

		// Update the UI with the selected path
		path := file.Path()
		glib.IdleAdd(func() {
			logger.Debug(logger.GUI, fmt.Sprintf("File selected: %s", path))
			if path != "" {
				sc.UI.Page4.VideoFileRow.SetSubtitle(path)
			}
		})
	}

	// Launch dialog
	fileDialog.Open(context.TODO(), &sc.UI.Window.Window, cb)

}

// saveSession handles the session Save/Save As... logic
func (sc *SessionController) saveSession(saveAs bool) {

	// Harvest the data from the UI widgets
	newConfig := sc.harvestEditor()
	currentPath := sc.SessionManager.ConfigPath()

	// Perform Save or Save As... action
	if saveAs || currentPath == "" {
		sc.openSaveAsDialog(newConfig)
	} else {
		glib.IdleAdd(func() {
			sc.performSessionSave(currentPath, newConfig)
		})
	}
}

// openSaveAsDialog handles the lifecycle of the file chooser
func (sc *SessionController) openSaveAsDialog(cfg *config.Config) {

	// Load the dialog if it's not already loaded
	if sc.saveFileDialog == nil {
		sc.saveFileDialog = gtk.NewFileDialog()
		sc.saveFileDialog.SetTitle("Save Session Configuration")
		sc.saveFileDialog.SetModal(true)
	}

	// Update dialog properties
	newFilename := fmt.Sprintf("%s.toml", convertSessionTitle(cfg.App.SessionTitle))
	sc.saveFileDialog.SetInitialName(newFilename)

	// Define the callback used to handle the file chooser
	cb := func(res gio.AsyncResulter) {

		file, err := sc.saveFileDialog.SaveFinish(res)
		if err != nil {
			return
		}

		filePath := file.Path()

		// Perform the actual save
		glib.IdleAdd(func() {
			sc.performSessionSave(filePath, cfg)
		})
	}

	// Launch the dialog
	sc.saveFileDialog.Save(context.TODO(), &sc.UI.Window.Window, cb)
}

// performSessionSave handles I/O, Error reporting, and UI Refresh
func (sc *SessionController) performSessionSave(path string, cfg *config.Config) {

	logger.Debug(logger.GUI, fmt.Sprintf("attempting to save session to: %s", path))

	// Perform file I/O
	if err := config.Save(path, cfg, config.GetVersion()); err != nil {
		logger.Error(logger.GUI, fmt.Sprintf("failed to save config: %v", err))
		displayAlertDialog(sc.UI.Window, "Save Error", err.Error())

		return
	}

	logger.Info(logger.GUI, fmt.Sprintf("session saved to %s", path))

	// Refresh the Session List (Page 1) to show new session(s)
	sc.scanForSessions()
	sc.PopulateSessionList()

}

// indexOf finds the index of a string in a slice (defaults to 0)
func indexOf(element string, data []string) uint {

	for k, v := range data {
		if element == v {
			return uint(k)
		}
	}

	return 0
}

// convertSessionTitle converts a session title into a string for use as a filename
func convertSessionTitle(sessionTitle string) string {

	if sessionTitle != "" {
		return strings.ReplaceAll(sessionTitle, " ", "_")
	}

	return "BSC_session"

}
