package ui

import (
	_ "embed" // required for go:embed
	"fmt"
	"os"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/flags"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
	"github.com/richbl/go-ble-sync-cycle/internal/services"
)

//go:embed assets/bsc_gui.ui
var uiXML string

// AppUI serves as the central controller for the GUI
type AppUI struct {
	Window      *adw.ApplicationWindow
	ViewStack   *adw.ViewStack
	Page1       *PageSessionSelect
	Page2       *PageSessionStatus
	Page3       *PageSessionLog
	Page4       *PageSessionEditor
	exitDialog  *adw.AlertDialog
	shutdownMgr *services.ShutdownManager
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
	SessionFileLocationRow   *adw.ActionRow
	SensorStatusRow          *adw.ActionRow
	SensorBatteryRow         *adw.ActionRow
	SpeedRow                 *adw.ActionRow
	SpeedLabel               *gtk.Label
	PlaybackSpeedRow         *adw.ActionRow
	PlaybackSpeedLabel       *gtk.Label
	TimeRemainingLabel       *gtk.Label
	TimeRemainingRow         *adw.ActionRow
	SessionControlRow        *adw.ActionRow
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

	// Scrolled window
	ScrolledWindow *adw.PreferencesPage

	// Session Details
	TitleEntry *adw.EntryRow
	LogLevel   *adw.ComboRow

	// BLE Sensor
	BTAddressEntry *adw.EntryRow
	ScanTimeout    *adw.SpinRow

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
	UpdateInterval  *adw.SpinRow
	SpeedMultiplier *adw.SpinRow

	// OSD
	SwitchCycleSpeed    *adw.SwitchRow
	SwitchPlaybackSpeed *adw.SwitchRow
	SwitchTimeRemaining *adw.SwitchRow
	FontSize            *adw.SpinRow
	MarginLeft          *adw.SpinRow
	MarginTop           *adw.SpinRow

	// Save Actions
	SaveRow      *adw.ActionRow
	SaveButton   *gtk.Button
	SaveAsButton *gtk.Button
}

// NewAppUI constructs the AppUI from the GTK-Builder GUI file (bsc_gui.ui)
func NewAppUI(builder *gtk.Builder) *AppUI {

	ui := &AppUI{
		Window:    objGTK[*adw.ApplicationWindow](builder, "main_window"),
		ViewStack: objGTK[*adw.ViewStack](builder, "view_stack"),
		Page1:     hydrateSessionSelect(builder),
		Page2:     hydrateSessionStatus(builder),
		Page3:     hydrateSessionLog(builder),
		Page4:     hydrateSessionEditor(builder),
	}

	return ui
}

// objGTK retrieves a GTK object by ID and casts it to type T
//
//nolint:ireturn // Generic function returning T must return the interface type
func objGTK[T any](builder *gtk.Builder, id string) T {

	obj := builder.GetObject(id)
	if obj == nil {
		logger.Fatal(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("widget ID '%s' not found in XML design file", id))
	}

	val, ok := obj.Cast().(T)
	if !ok {
		logger.Fatal(logger.BackgroundCtx, logger.GUI, fmt.Sprintf("widget ID '%s' is not of expected type %T", id, *new(T)))
	}

	return val
}

// hydrateSessionSelect constructs the PageSessionSelect from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionSelect(builder *gtk.Builder) *PageSessionSelect {

	return &PageSessionSelect{
		ListBox:    objGTK[*gtk.ListBox](builder, "session_listbox"),
		EditButton: objGTK[*gtk.Button](builder, "edit_session_button"),
		LoadButton: objGTK[*gtk.Button](builder, "load_session_button"),
	}
}

// hydrateSessionStatus constructs the PageSessionStatus from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionStatus(builder *gtk.Builder) *PageSessionStatus {

	return &PageSessionStatus{
		SessionNameRow:           objGTK[*adw.ActionRow](builder, "session_name_row"),
		SessionFileLocationRow:   objGTK[*adw.ActionRow](builder, "session_file_location_row"),
		SensorStatusRow:          objGTK[*adw.ActionRow](builder, "sensor_status_row"),
		SensorBatteryRow:         objGTK[*adw.ActionRow](builder, "battery_level_row"),
		SpeedRow:                 objGTK[*adw.ActionRow](builder, "speed_row"),
		SpeedLabel:               objGTK[*gtk.Label](builder, "speed_large_label"),
		PlaybackSpeedLabel:       objGTK[*gtk.Label](builder, "playback_speed_large_label"),
		PlaybackSpeedRow:         objGTK[*adw.ActionRow](builder, "playback_speed_row"),
		TimeRemainingLabel:       objGTK[*gtk.Label](builder, "time_remaining_large_label"),
		TimeRemainingRow:         objGTK[*adw.ActionRow](builder, "time_remaining_row"),
		SessionControlRow:        objGTK[*adw.ActionRow](builder, "session_control_row"),
		SessionControlBtn:        objGTK[*gtk.Button](builder, "session_control_button"),
		SessionControlBtnContent: objGTK[*adw.ButtonContent](builder, "session_control_button_content"),
		SensorConnIcon:           objGTK[*gtk.Image](builder, "connection_status_icon"),
		SensorBattIcon:           objGTK[*gtk.Image](builder, "battery_icon"),
	}
}

// hydrateSessionLog constructs the PageSessionLog from the GTK-Builder GUI file (bsc_gui.ui)
func hydrateSessionLog(builder *gtk.Builder) *PageSessionLog {

	sessionLog := &PageSessionLog{
		LogLevelRow: objGTK[*adw.ActionRow](builder, "logging_level_row"),
		TextView:    objGTK[*gtk.TextView](builder, "logging_view"),
	}

	// Display logging level
	sessionLog.LogLevelRow.SetTitle(logger.LogLevel())

	// Configure TextView for logging
	tv := sessionLog.TextView
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
	scrolledWindow := objGTK[*gtk.ScrolledWindow](builder, "logging_scroll_window")
	vAdj := scrolledWindow.VAdjustment()
	vAdj.Connect("changed", func() {

		// Calculate the bottom-most position
		target := vAdj.Upper() - vAdj.PageSize()

		// Only scroll if there is scrollable content
		if target > 0 {
			vAdj.SetValue(target)
		}

	})

	// Set up logging bridge (permits logger GUI output)
	sessionLog.LogWriter = NewGuiLogWriter(tv)

	// Enable logging to the console in GUI mode if requested
	if flags.IsGUIConsoleLogging() {
		logger.AddWriter(sessionLog.LogWriter)
		logger.Info(logger.BackgroundCtx, logger.GUI, "logging via Session Log started with added console/CLI output")
	} else {
		logger.UseGUIWriterOnly(sessionLog.LogWriter)
		logger.Info(logger.BackgroundCtx, logger.GUI, "logging via Session Log started")
	}

	return sessionLog
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
		ScrolledWindow:      objGTK[*adw.PreferencesPage](builder, "session_editor_page"),
		TitleEntry:          objGTK[*adw.EntryRow](builder, "session_title_entry_row"),
		LogLevel:            objGTK[*adw.ComboRow](builder, "log_level_combo"),
		BTAddressEntry:      objGTK[*adw.EntryRow](builder, "bt_address_entry_row"),
		ScanTimeout:         objGTK[*adw.SpinRow](builder, "scan_timeout_spin"),
		WheelCircumference:  objGTK[*adw.SpinRow](builder, "edit_wheel_circumference_spin"),
		SpeedUnits:          objGTK[*adw.ComboRow](builder, "edit_speed_units_combo"),
		SpeedThreshold:      objGTK[*adw.SpinRow](builder, "edit_speed_threshold_spin"),
		SpeedSmoothing:      objGTK[*adw.SpinRow](builder, "edit_speed_smoothing_spin"),
		MediaPlayer:         objGTK[*adw.ComboRow](builder, "edit_media_player_combo"),
		VideoFileRow:        objGTK[*adw.ActionRow](builder, "video_file_row"),
		VideoFileButton:     objGTK[*gtk.Button](builder, "video_file_button"),
		StartTimeEntry:      objGTK[*adw.EntryRow](builder, "start_time_entry_row"),
		WindowScale:         objGTK[*adw.SpinRow](builder, "edit_window_scale_factor_spin"),
		UpdateInterval:      objGTK[*adw.SpinRow](builder, "edit_update_interval_spin"),
		SpeedMultiplier:     objGTK[*adw.SpinRow](builder, "edit_speed_multiplier_spin"),
		SwitchCycleSpeed:    objGTK[*adw.SwitchRow](builder, "display_cycle_speed_switch"),
		SwitchPlaybackSpeed: objGTK[*adw.SwitchRow](builder, "display_playback_speed_switch"),
		SwitchTimeRemaining: objGTK[*adw.SwitchRow](builder, "display_time_remaining_switch"),
		FontSize:            objGTK[*adw.SpinRow](builder, "display_font_size_spin"),
		MarginLeft:          objGTK[*adw.SpinRow](builder, "pixel_offset_left_spin"),
		MarginTop:           objGTK[*adw.SpinRow](builder, "pixel_offset_top_spin"),
		SaveRow:             objGTK[*adw.ActionRow](builder, "edit_save_row"),
		SaveButton:          objGTK[*gtk.Button](builder, "save_button"),
		SaveAsButton:        objGTK[*gtk.Button](builder, "save_as_button"),
	}
}

// setupAllSignals sets up all UI signal handlers for the application
func setupAllSignals(sc *SessionController) {

	// Generalized navigation setup (handles all pages via map)
	pageActions := map[string]func(){

		"page1": func() {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "view switched to Session Select: refreshing session list from CWD...")
			sc.scanForSessions()
			sc.PopulateSessionList()
		},

		"page2": func() {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "view switched to Session Status")
		},

		"page3": func() {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "view switched to Session Log")
			sc.UpdateLogLevel()
		},

		"page4": func() {
			logger.Debug(logger.BackgroundCtx, logger.GUI, "view switched to Session Editor")
			sc.UI.Page4.ScrolledWindow.ScrollToTop()
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

	// Create a global ShutdownManager for GUI mode to handle signals (CTRL+C)
	logger.Debug(logger.BackgroundCtx, logger.GUI, "creating GUI ShutdownManager for signal handling")
	shutdownMgr := services.NewShutdownManager(30 * time.Second)

	// Initialize the application
	app := gtk.NewApplication("com.github.richbl.ble-sync-cycle", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		setupGUIApplication(app, shutdownMgr)
	})

	// Set up signal handling for CTRL+C that integrates with GTK event loop
	logger.Debug(logger.BackgroundCtx, logger.GUI, "starting ShutdownManager signal handler")
	shutdownMgr.Start()

	// Monitor shutdown signal in a goroutine and trigger GTK quit when signaled
	go func() {
		ctx := *shutdownMgr.Context()
		<-ctx.Done()
		fmt.Fprint(os.Stdout, "\r") // Clear the ^C character from the terminal line
		logger.Info(logger.BackgroundCtx, logger.GUI, "shutdown signal received, triggering GTK application quit")

		// Use glib.IdleAdd to safely quit the GTK application from the main thread
		glib.IdleAdd(func() {
			app.Quit()
		})
	}()

	// Run the GUI application... fly and be free!
	logger.Debug(logger.BackgroundCtx, logger.GUI, "starting GTK event loop")
	app.Run(nil)

	// Application has exited, perform cleanup
	logger.Debug(logger.BackgroundCtx, logger.GUI, "GTK event loop exited, performing cleanup")
	shutdownMgr.Shutdown()
	services.WaveGoodbye(logger.BackgroundCtx)

}
