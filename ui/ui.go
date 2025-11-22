package ui

import (
	_ "embed"
	"log"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Embed the UI XML file into the binary
//
//go:embed assets/bsc_gui.ui
var uiXML string

// AppUI serves as the central controller for the GUI.
type AppUI struct {
	Window    *adw.ApplicationWindow
	ViewStack *adw.ViewStack

	Page1 *PageSessionSelect
	Page2 *PageSessionStatus
	Page3 *PageSessionLog
	Page4 *PageSessionEditor
}

type PageSessionSelect struct {
	ListBox    *gtk.ListBox
	EditButton *gtk.Button
	LoadButton *gtk.Button
}

type PageSessionStatus struct {
	SessionNameRow     *adw.ActionRow
	SensorStatusRow    *adw.ActionRow
	SensorBatteryRow   *adw.ActionRow
	SpeedLabel         *gtk.Label
	PlaybackSpeedLabel *gtk.Label
	TimeRemainingLabel *gtk.Label
	SessionControlBtn  *gtk.Button
	SensorConnIcon     *gtk.Image
	SensorBattIcon     *gtk.Image
}

type PageSessionLog struct {
	LogLevelRow *adw.ActionRow
	TextView    *gtk.TextView
}

type PageSessionEditor struct {
	// Session Details
	TitleEntry *adw.EntryRow
	LogLevel   *adw.ComboRow

	// BLE Sensor
	BTAddressEntry *adw.EntryRow
	ScanTimeout    *adw.SpinRow
	SensorStatus   *adw.ActionRow
	ScanButton     *gtk.Button

	// Speed Settings
	WheelCircumference *adw.SpinRow
	SpeedUnits         *adw.ComboRow
	SpeedThreshold     *adw.SpinRow
	SpeedSmoothing     *adw.SpinRow

	// Video Settings
	MediaPlayer     *adw.ComboRow
	VideoFileRow    *adw.ActionRow
	VideoFileButton *gtk.Button
	StartTimeEntry  *adw.EntryRow
	WindowScale     *adw.SpinRow
	SpeedMultiplier *adw.SpinRow

	// OSD
	SwitchCycleSpeed    *adw.SwitchRow
	SwitchPlaybackSpeed *adw.SwitchRow
	SwitchTimeRemaining *adw.SwitchRow
	FontSize            *adw.SpinRow
	MarginLeft          *adw.SpinRow
	MarginTop           *adw.SpinRow

	// Save Actions
	SaveButton   *gtk.Button
	SaveAsButton *gtk.Button
}

// NewAppUI initializes the AppUI structure by retrieving widgets from the GTK builder
func NewAppUI(builder *gtk.Builder) *AppUI {

	ui := &AppUI{
		Window:    getObj(builder, "main_window").Cast().(*adw.ApplicationWindow),
		ViewStack: getObj(builder, "view_stack").Cast().(*adw.ViewStack),
		Page1:     hydrateSessionSelect(builder),
		Page2:     hydrateSessionStatus(builder),
		Page3:     hydrateSessionLog(builder),
		Page4:     hydrateSessionEditor(builder),
	}

	return ui
}

// getObj is a helper function to retrieve an object from the builder and ensure it exists
func getObj(builder *gtk.Builder, id string) *glib.Object {

	obj := builder.GetObject(id)
	if obj == nil {
		log.Fatalf("Critical: Widget ID '%s' not found in XML.", id)
	}

	return obj
}

// hydrateSessionSelect populates the PageSessionSelect structure with widgets from the builder
func hydrateSessionSelect(builder *gtk.Builder) *PageSessionSelect {

	return &PageSessionSelect{
		ListBox:    getObj(builder, "session_listbox").Cast().(*gtk.ListBox),
		EditButton: getObj(builder, "edit_session_button").Cast().(*gtk.Button),
		LoadButton: getObj(builder, "load_session_button").Cast().(*gtk.Button),
	}
}

// hydrateSessionStatus populates the PageSessionStatus structure with widgets from the builder
func hydrateSessionStatus(builder *gtk.Builder) *PageSessionStatus {

	return &PageSessionStatus{
		SessionNameRow:     getObj(builder, "session_name_row").Cast().(*adw.ActionRow),
		SensorStatusRow:    getObj(builder, "sensor_status_row").Cast().(*adw.ActionRow),
		SensorBatteryRow:   getObj(builder, "battery_level_row").Cast().(*adw.ActionRow),
		SpeedLabel:         getObj(builder, "speed_large_label").Cast().(*gtk.Label),
		PlaybackSpeedLabel: getObj(builder, "playback_speed_large_label").Cast().(*gtk.Label),
		TimeRemainingLabel: getObj(builder, "time_remaining_large_label").Cast().(*gtk.Label),
		SessionControlBtn:  getObj(builder, "session_control_button").Cast().(*gtk.Button),
		SensorConnIcon:     getObj(builder, "connection_status_icon").Cast().(*gtk.Image),
		SensorBattIcon:     getObj(builder, "battery_icon").Cast().(*gtk.Image),
	}
}

// hydrateSessionLog populates the PageSessionLog structure with widgets from the builder
func hydrateSessionLog(builder *gtk.Builder) *PageSessionLog {

	return &PageSessionLog{
		LogLevelRow: getObj(builder, "logging_level_row").Cast().(*adw.ActionRow),
		TextView:    getObj(builder, "logging_view").Cast().(*gtk.TextView),
	}
}

// hydrateSessionEditor populates the PageSessionEditor structure with widgets from the builder
func hydrateSessionEditor(builder *gtk.Builder) *PageSessionEditor {

	return &PageSessionEditor{
		TitleEntry:          getObj(builder, "session_title_entry_row").Cast().(*adw.EntryRow),
		LogLevel:            getObj(builder, "log_level_combo").Cast().(*adw.ComboRow),
		BTAddressEntry:      getObj(builder, "bt_address_entry_row").Cast().(*adw.EntryRow),
		ScanTimeout:         getObj(builder, "scan_timeout_spin").Cast().(*adw.SpinRow),
		SensorStatus:        getObj(builder, "edit_sensor_status_row").Cast().(*adw.ActionRow),
		ScanButton:          getObj(builder, "button_scan").Cast().(*gtk.Button),
		WheelCircumference:  getObj(builder, "edit_wheel_circumference_spin").Cast().(*adw.SpinRow),
		SpeedUnits:          getObj(builder, "edit_speed_units_combo").Cast().(*adw.ComboRow),
		SpeedThreshold:      getObj(builder, "edit_speed_threshold_spin").Cast().(*adw.SpinRow),
		SpeedSmoothing:      getObj(builder, "edit_speed_smoothing_spin").Cast().(*adw.SpinRow),
		MediaPlayer:         getObj(builder, "edit_media_player_combo").Cast().(*adw.ComboRow),
		VideoFileRow:        getObj(builder, "video_file_row").Cast().(*adw.ActionRow),
		VideoFileButton:     getObj(builder, "video_file_button").Cast().(*gtk.Button),
		StartTimeEntry:      getObj(builder, "start_time_entry_row").Cast().(*adw.EntryRow),
		WindowScale:         getObj(builder, "edit_window_scale_factor_spin").Cast().(*adw.SpinRow),
		SpeedMultiplier:     getObj(builder, "edit_speed_multiplier_spin").Cast().(*adw.SpinRow),
		SwitchCycleSpeed:    getObj(builder, "display_cycle_speed_switch").Cast().(*adw.SwitchRow),
		SwitchPlaybackSpeed: getObj(builder, "display_playback_speed_switch").Cast().(*adw.SwitchRow),
		SwitchTimeRemaining: getObj(builder, "display_time_remaining_switch").Cast().(*adw.SwitchRow),
		FontSize:            getObj(builder, "display_font_size_spin").Cast().(*adw.SpinRow),
		MarginLeft:          getObj(builder, "pixel_offset_left_spin").Cast().(*adw.SpinRow),
		MarginTop:           getObj(builder, "pixel_offset_top_spin").Cast().(*adw.SpinRow),
		SaveButton:          getObj(builder, "save_button").Cast().(*gtk.Button),
		SaveAsButton:        getObj(builder, "save_as_button").Cast().(*gtk.Button),
	}
}

// StartGUI initializes and runs the GTK application
func StartGUI() {
	app := gtk.NewApplication("com.github.richbl.ble-sync-cycle", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		adw.Init()

		builder := gtk.NewBuilderFromString(uiXML)
		ui := NewAppUI(builder)

		aboutWindow := get(builder, "about_window").Cast().(*adw.AboutDialog)

		// Create the "About" action handler
		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(_ *glib.Variant) {

			var transientParent gtk.Widgetter = ui.Window
			aboutWindow.Present(transientParent)
		})

		app.AddAction(aboutAction)

		sessionCtrl := NewSessionController(ui)
		sessionCtrl.scanForSessions()
		sessionCtrl.PopulateSessionList()
		sessionCtrl.SetupSignals()

		ui.Window.SetApplication(app)
		ui.Window.Present()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}

}

// get is a helper function to retrieve an object from the builder and ensure it exists
func get(builder *gtk.Builder, id string) *glib.Object {

	obj := builder.GetObject(id)
	if obj == nil {
		log.Fatalf("Critical: Widget ID '%s' not found in XML.", id)
	}

	return obj
}
