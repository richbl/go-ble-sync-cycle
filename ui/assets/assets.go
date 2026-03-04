package assets

import "embed"

// InstallerAssets embeds the .desktop and .svg icon files to consolidate the installation process
//
//go:embed com.github.richbl.ble-sync-cycle.desktop com.github.richbl.ble-sync-cycle.svg
var InstallerAssets embed.FS
