package ui

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
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

// PageSessionSelect holds widgets for the Session Selection tab (Page 1)
type PageSessionSelect struct {
	ListBox    *gtk.ListBox
	EditButton *gtk.Button
	LoadButton *gtk.Button
}

// PageSessionStatus holds widgets for the Session Status tab (Page 2)
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

// PageSessionLog holds widgets for the Session Log tab (Page 3)
type PageSessionLog struct {
	LogLevelRow *adw.ActionRow
	TextView    *gtk.TextView
	LogWriter   *GuiLogWriter
}

// PageSessionEditor holds widgets for the Session Edit tab (Page 4)
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

// NewAppUI constructs the AppUI from the GTK-Builder GUI file (bsc_gui.ui)
func NewAppUI(builder *gtk.Builder) *AppUI {

	ui := &AppUI{
		Window:    objGTK(builder, "main_window").Cast().(*adw.ApplicationWindow),
		ViewStack: objGTK(builder, "view_stack").Cast().(*adw.ViewStack),
		Page1:     hydrateSessionSelect(builder),
		Page2:     hydrateSessionStatus(builder),
		Page3:     hydrateSessionLog(builder),
		Page4:     hydrateSessionEditor(builder),
	}

	return ui
}

// objGTK retrieves a GTK object by ID and logs a fatal error if not found
func objGTK(builder *gtk.Builder, id string) *glib.Object {

	obj := builder.GetObject(id)
	if obj == nil {
		logger.Fatal(logger.GUI, fmt.Sprintf("Widget ID '%s' not found in XML design file", id))
	}

	return obj
}

// hydrateSessionSelect constructs the PageSessionSelect from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionSelect(builder *gtk.Builder) *PageSessionSelect {

	return &PageSessionSelect{
		ListBox:    objGTK(builder, "session_listbox").Cast().(*gtk.ListBox),
		EditButton: objGTK(builder, "edit_session_button").Cast().(*gtk.Button),
		LoadButton: objGTK(builder, "load_session_button").Cast().(*gtk.Button),
	}
}

// hydrateSessionStatus constructs the PageSessionStatus from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionStatus(builder *gtk.Builder) *PageSessionStatus {

	return &PageSessionStatus{
		SessionNameRow:           objGTK(builder, "session_name_row").Cast().(*adw.ActionRow),
		SensorStatusRow:          objGTK(builder, "sensor_status_row").Cast().(*adw.ActionRow),
		SensorBatteryRow:         objGTK(builder, "battery_level_row").Cast().(*adw.ActionRow),
		SpeedLabel:               objGTK(builder, "speed_large_label").Cast().(*gtk.Label),
		PlaybackSpeedLabel:       objGTK(builder, "playback_speed_large_label").Cast().(*gtk.Label),
		TimeRemainingLabel:       objGTK(builder, "time_remaining_large_label").Cast().(*gtk.Label),
		SessionControlBtn:        objGTK(builder, "session_control_button").Cast().(*gtk.Button),
		SessionControlBtnContent: objGTK(builder, "session_control_button_content").Cast().(*adw.ButtonContent),
		SensorConnIcon:           objGTK(builder, "connection_status_icon").Cast().(*gtk.Image),
		SensorBattIcon:           objGTK(builder, "battery_icon").Cast().(*gtk.Image),
	}
}

// hydrateSessionLog constructs the PageSessionLog from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionLog(builder *gtk.Builder) *PageSessionLog {

	page := &PageSessionLog{
		LogLevelRow: objGTK(builder, "logging_level_row").Cast().(*adw.ActionRow),
		TextView:    objGTK(builder, "logging_view").Cast().(*gtk.TextView),
	}

	// Display logging level
	page.LogLevelRow.SetTitle(logger.LogLevel())

	// Configure TextView for logging
	tv := page.TextView
	tv.SetMonospace(true)
	tv.SetEditable(false)
	tv.SetCursorVisible(false)
	tv.SetLeftMargin(10)
	tv.SetRightMargin(10)
	tv.SetTopMargin(10)
	tv.SetBottomMargin(10)
	tv.SetWrapMode(gtk.WrapNone)

	tv.AddCSSClass("session-log-view")
	applyLogStyles()

	// Configure scrolling behavior to always scroll to the bottom
	scrolledWindow := objGTK(builder, "logging_scroll_window").Cast().(*gtk.ScrolledWindow)
	vAdj := scrolledWindow.VAdjustment()
	vAdj.Connect("changed", func() {

		// Calculate the bottom-most position
		target := vAdj.Upper() - vAdj.PageSize()

		// Only scroll if there is scrollable content
		if target > 0 {
			vAdj.SetValue(target)
		}

	})

	// Set up logging bridge
	page.LogWriter = NewGuiLogWriter(tv)
	logger.AddWriter(page.LogWriter)

	logger.Info(logger.GUI, "Session Log bridge established")

	return page
}

// applyLogStyles injects a CSS provider to style the log view specifically
func applyLogStyles() {

	// Create CSS styles that define the log view
	css := `
	.session-log-view {
		font-size: 11px;
		font-family: 'Monospace';
	}
	`
	provider := gtk.NewCSSProvider()
	provider.LoadFromString(css)

	display := gdk.DisplayGetDefault()
	if display != nil {
		gtk.StyleContextAddProviderForDisplay(display, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	}

}

// hydrateSessionEditor constructs the PageSessionEditor from the GTK builder
func hydrateSessionEditor(builder *gtk.Builder) *PageSessionEditor {

	return &PageSessionEditor{
		TitleEntry:          objGTK(builder, "session_title_entry_row").Cast().(*adw.EntryRow),
		LogLevel:            objGTK(builder, "log_level_combo").Cast().(*adw.ComboRow),
		BTAddressEntry:      objGTK(builder, "bt_address_entry_row").Cast().(*adw.EntryRow),
		ScanTimeout:         objGTK(builder, "scan_timeout_spin").Cast().(*adw.SpinRow),
		WheelCircumference:  objGTK(builder, "edit_wheel_circumference_spin").Cast().(*adw.SpinRow),
		SpeedUnits:          objGTK(builder, "edit_speed_units_combo").Cast().(*adw.ComboRow),
		SpeedThreshold:      objGTK(builder, "edit_speed_threshold_spin").Cast().(*adw.SpinRow),
		SpeedSmoothing:      objGTK(builder, "edit_speed_smoothing_spin").Cast().(*adw.SpinRow),
		MediaPlayer:         objGTK(builder, "edit_media_player_combo").Cast().(*adw.ComboRow),
		VideoFileRow:        objGTK(builder, "video_file_row").Cast().(*adw.ActionRow),
		VideoFileButton:     objGTK(builder, "video_file_button").Cast().(*gtk.Button),
		StartTimeEntry:      objGTK(builder, "start_time_entry_row").Cast().(*adw.EntryRow),
		WindowScale:         objGTK(builder, "edit_window_scale_factor_spin").Cast().(*adw.SpinRow),
		SpeedMultiplier:     objGTK(builder, "edit_speed_multiplier_spin").Cast().(*adw.SpinRow),
		SwitchCycleSpeed:    objGTK(builder, "display_cycle_speed_switch").Cast().(*adw.SwitchRow),
		SwitchPlaybackSpeed: objGTK(builder, "display_playback_speed_switch").Cast().(*adw.SwitchRow),
		SwitchTimeRemaining: objGTK(builder, "display_time_remaining_switch").Cast().(*adw.SwitchRow),
		FontSize:            objGTK(builder, "display_font_size_spin").Cast().(*adw.SpinRow),
		MarginLeft:          objGTK(builder, "pixel_offset_left_spin").Cast().(*adw.SpinRow),
		MarginTop:           objGTK(builder, "pixel_offset_top_spin").Cast().(*adw.SpinRow),
		SaveButton:          objGTK(builder, "save_button").Cast().(*gtk.Button),
		SaveAsButton:        objGTK(builder, "save_as_button").Cast().(*gtk.Button),
	}
}

// setupAllSignals sets up all UI signal handlers for the application
func setupAllSignals(sc *SessionController) {

	// Generalized navigation setup (handles all pages via map)
	pageActions := map[string]func(){

		"page1": func() {
			logger.Debug(logger.GUI, "view switched to Session Select: refreshing session list from CWD...")
			sc.scanForSessions()
			sc.PopulateSessionList()
		},

		"page2": func() {
			logger.Debug(logger.GUI, "view switched to Session Status")
		},

		"page3": func() {
			logger.Debug(logger.GUI, "view switched to Session Log")
			sc.UpdateLogLevel()
		},

		"page4": func() {
			logger.Debug(logger.GUI, "view switched to Session Editor")
		},
	}

	// Reuse existing navigation setup utility
	setupNavigationSignals(sc.UI.ViewStack, pageActions)

	// Per-tab signal setups
	sc.setupSessionSelectSignals()
	sc.setupSessionStatusSignals()
	sc.setupSessionLogSignals()
	sc.setupSessionEditSignals()

}

// StartGUI initializes and runs the GTK4/Adwaita application
func StartGUI() {

	// Initialize the application
	app := gtk.NewApplication("com.github.richbl.ble-sync-cycle", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {

		adw.Init()
		builder := gtk.NewBuilderFromString(uiXML)
		ui := NewAppUI(builder)
		aboutWindow := objGTK(builder, "about_window").Cast().(*adw.AboutDialog)

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

	// Run the GUI application... fly and be free!
	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}

}
