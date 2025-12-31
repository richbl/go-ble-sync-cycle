package ui

// Session represents the configuration file and its display name
type Session struct {
	ID         int
	Title      string
	ConfigPath string
}

// Status represents the logical connection/battery status
type Status int

const (
	StatusConnected Status = iota
	StatusNotConnected
	StatusStopped
	StatusConnecting
	StatusFailed
)

// ObjectKind represents the logical object we are displaying
type ObjectKind int

const (
	ObjectBLE ObjectKind = iota
	ObjectBattery
)

// StatusPresentation holds the UI-facing data for a status
type StatusPresentation struct {
	Display  string
	Icon     string
	CSSStyle string
}

const (
	// BLE icons
	iconBLEConnected    = "bluetooth-symbolic"
	iconBLENotConnected = "bluetooth-disconnected-symbolic"
	iconBLEConnecting   = "bluetooth-acquiring-symbolic"

	// Battery icons
	iconBatteryConnected    = "battery-good-symbolic"
	iconBatteryNotConnected = "battery-symbolic"
	iconBatteryConnecting   = "battery-symbolic"
)

// statusTable centralizes all mappings of (object, status, style/color) -> UI data
var statusTable = map[ObjectKind]map[Status]StatusPresentation{
	ObjectBLE: {
		StatusConnected:    {Display: "Connected", Icon: iconBLEConnected, CSSStyle: "success"},
		StatusNotConnected: {Display: "Not Connected", Icon: iconBLENotConnected, CSSStyle: "error"},
		StatusStopped:      {Display: "Stopped", Icon: iconBLENotConnected, CSSStyle: "error"},
		StatusConnecting:   {Display: "Connecting...", Icon: iconBLEConnecting, CSSStyle: "warning"},
		StatusFailed:       {Display: "Failed", Icon: iconBLENotConnected, CSSStyle: "error"},
	},
	ObjectBattery: {
		StatusConnected:    {Display: "Connected", Icon: iconBatteryConnected, CSSStyle: "success"},
		StatusNotConnected: {Display: "Unknown", Icon: iconBatteryNotConnected, CSSStyle: "error"},
		StatusStopped:      {Display: "Unknown", Icon: iconBatteryNotConnected, CSSStyle: "error"},
		StatusConnecting:   {Display: "Connecting...", Icon: iconBatteryConnecting, CSSStyle: "warning"},
		StatusFailed:       {Display: "Unknown", Icon: iconBatteryNotConnected, CSSStyle: "error"},
	},
}
