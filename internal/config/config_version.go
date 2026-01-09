package config

// Application name and version information
const (
	appName    = "BLE Sync Cycle"
	appVersion = "0.50.0"
)

// GetVersion returns the current application version
func GetVersion() string {
	return appVersion
}

// GetAppName returns the application name
func GetAppName() string {
	return appName
}

// GetFullVersion returns the app name and version combined
func GetFullVersion() string {
	return appName + " v" + appVersion
}
