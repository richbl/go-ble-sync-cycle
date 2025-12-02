package ui

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed assets/bsc_gui.ui
var uiXML string

// AppUI serves as the central controller for the GUI
type AppUI struct {
	Window    *adw.ApplicationWindow
	ViewStack *adw.ViewStack
	Page1     *PageSessionSelect
	Page2     *PageSessionStatus
	Page3     *PageSessionLog
	Page4     *PageSessionEditor
}

// PageSessionSelect holds widgets for the session selection tab (Page 1)
type PageSessionSelect struct {
	ListBox    *gtk.ListBox
	EditButton *gtk.Button
	LoadButton *gtk.Button
}

// PageSessionStatus holds widgets for the session status tab (Page 2)
type PageSessionStatus struct {
	SessionNameRow           *adw.ActionRow
	SensorStatusRow          *adw.ActionRow
	SensorBatteryRow         *adw.ActionRow
	SpeedLabel               *gtk.Label
	PlaybackSpeedLabel       *gtk.Label
	TimeRemainingLabel       *gtk.Label
	SessionControlBtn        *gtk.Button
	SessionControlBtnContent *adw.ButtonContent
	SensorConnIcon           *gtk.Image
	SensorBattIcon           *gtk.Image
}

// PageSessionLog holds widgets for the logging tab (Page 3)
type PageSessionLog struct {
	LogLevelRow *adw.ActionRow
	TextView    *gtk.TextView
}

// PageSessionEditor holds widgets for the session editor tab (Page 4)
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

// NewAppUI constructs the AppUI from the GTK builder
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

// getObj retrieves a GTK object by ID and logs a fatal error if not found
func getObj(builder *gtk.Builder, id string) *glib.Object {

	obj := builder.GetObject(id)
	if obj == nil {
		log.Fatalf("Critical: Widget ID '%s' not found in XML.", id)
	}

	return obj
}

// hydrateSessionSelect constructs the PageSessionSelect from the GTK builder
func hydrateSessionSelect(builder *gtk.Builder) *PageSessionSelect {

	return &PageSessionSelect{
		ListBox:    getObj(builder, "session_listbox").Cast().(*gtk.ListBox),
		EditButton: getObj(builder, "edit_session_button").Cast().(*gtk.Button),
		LoadButton: getObj(builder, "load_session_button").Cast().(*gtk.Button),
	}
}

// hydrateSessionStatus constructs the PageSessionStatus from the GTK builder
func hydrateSessionStatus(builder *gtk.Builder) *PageSessionStatus {

	return &PageSessionStatus{
		SessionNameRow:           getObj(builder, "session_name_row").Cast().(*adw.ActionRow),
		SensorStatusRow:          getObj(builder, "sensor_status_row").Cast().(*adw.ActionRow),
		SensorBatteryRow:         getObj(builder, "battery_level_row").Cast().(*adw.ActionRow),
		SpeedLabel:               getObj(builder, "speed_large_label").Cast().(*gtk.Label),
		PlaybackSpeedLabel:       getObj(builder, "playback_speed_large_label").Cast().(*gtk.Label),
		TimeRemainingLabel:       getObj(builder, "time_remaining_large_label").Cast().(*gtk.Label),
		SessionControlBtn:        getObj(builder, "session_control_button").Cast().(*gtk.Button),
		SessionControlBtnContent: getObj(builder, "session_control_button_content").Cast().(*adw.ButtonContent),
		SensorConnIcon:           getObj(builder, "connection_status_icon").Cast().(*gtk.Image),
		SensorBattIcon:           getObj(builder, "battery_icon").Cast().(*gtk.Image),
	}

}

// hydrateSessionLog constructs the PageSessionLog from the GTK builder
func hydrateSessionLog(builder *gtk.Builder) *PageSessionLog {

	return &PageSessionLog{
		LogLevelRow: getObj(builder, "logging_level_row").Cast().(*adw.ActionRow),
		TextView:    getObj(builder, "logging_view").Cast().(*gtk.TextView),
	}

}

// hydrateSessionEditor constructs the PageSessionEditor from the GTK builder
func hydrateSessionEditor(builder *gtk.Builder) *PageSessionEditor {

	return &PageSessionEditor{
		TitleEntry:          getObj(builder, "session_title_entry_row").Cast().(*adw.EntryRow),
		LogLevel:            getObj(builder, "log_level_combo").Cast().(*adw.ComboRow),
		BTAddressEntry:      getObj(builder, "bt_address_entry_row").Cast().(*adw.EntryRow),
		ScanTimeout:         getObj(builder, "scan_timeout_spin").Cast().(*adw.SpinRow),
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

// setupAllSignals sets up all UI signal handlers for the application
func setupAllSignals(sc *SessionController) {

	// Generalized navigation setup (handles all pages via map)
	pageActions := map[string]func(){

		"page1": func() {
			fmt.Println("View switched to Page 1: Refreshing Session List from CWD...")
			sc.scanForSessions()
			sc.PopulateSessionList()
		},
		// "page2": func() { /* e.g., refresh metrics */ },
		// "page3": func() { /* e.g., scroll to bottom of logs */ },
		"page4": func() {
			fmt.Println("Entered Editor (page 4): Load session config into fields (stubbed)")
			// TODO: Populate sc.UI.Page4 fields from sc.SessionManager.GetConfig()
		},
	}

	// Reuse existing navigation setup utility
	setupNavigationSignals(sc.UI.ViewStack, pageActions)

	// Per-tab signal setups
	sc.setupPage1Signals()
	sc.setupPage2Signals()
	sc.setupLogsSignals()
	sc.setupEditSignals()

}

// StartGUI initializes and runs the GTK application
func StartGUI() {

	app := gtk.NewApplication("com.github.richbl.ble-sync-cycle", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {

		adw.Init()
		builder := gtk.NewBuilderFromString(uiXML)
		ui := NewAppUI(builder)
		aboutWindow := getObj(builder, "about_window").Cast().(*adw.AboutDialog)

		// Create the "About" action handler
		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(_ *glib.Variant) {
			var transientParent gtk.Widgetter = ui.Window
			aboutWindow.Present(transientParent)
		})

		app.AddAction(aboutAction)

		// Create SessionController and initialize
		sessionCtrl := NewSessionController(ui)
		sessionCtrl.scanForSessions()
		sessionCtrl.PopulateSessionList()

		setupAllSignals(sessionCtrl)
		ui.Window.SetApplication(app)
		ui.Window.Present()
	})

	// Run the GUI application
	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}

}
