// Package installer provides a self-contained installation procedure for BLE Sync Cycle (BSC)
//
// When the user runs the BSC binary with the --install (or -i) flag, this package handles the
// installation, or uninstallation with the --uninstall (or -u flag), of the application and
// its associated assets into the user's local directories:
//
// - The BSC binary is copied to $XDG_BIN_HOME (default: ~/.local/bin)
// - The .desktop file is copied to $XDG_DATA_HOME/applications (default: ~/.local/share/applications)
// - The .svg icon is copied to $XDG_DATA_HOME/icons/hicolor/scalable/apps (default: ~/.local/share/icons/hicolor/scalable/apps)
//
// This allows users to easily install BSC without needing to manually move files or create
// desktop entries
package installer
