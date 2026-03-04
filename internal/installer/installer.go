package installer

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/richbl/go-ble-sync-cycle/internal/config"
	"github.com/richbl/go-ble-sync-cycle/ui/assets"
)

const (
	binFilename     = "ble-sync-cycle"
	desktopFilename = "com.github.richbl.ble-sync-cycle.desktop"
	iconFilename    = "com.github.richbl.ble-sync-cycle.svg"
)

var (
	errInvalidAppDir = errors.New("invalid application directory")
)

// installPaths holds all the relevant paths for installation/uninstallation
type installPaths struct {
	binDir        string
	appDir        string
	iconDir       string
	binPath       string
	desktopPath   string
	iconPath      string
	installAction bool
}

// newInstallPath resolves all necessary asset installation paths
func newInstallPath() (*installPaths, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	binDir := getBinHome(homeDir)
	dataDir := getDataHome(homeDir)
	appDir := filepath.Join(dataDir, "applications")
	iconDir := filepath.Join(dataDir, "icons", "hicolor", "scalable", "apps")

	return &installPaths{
		binDir:        binDir,
		appDir:        appDir,
		iconDir:       iconDir,
		binPath:       filepath.Join(binDir, binFilename),
		desktopPath:   filepath.Join(appDir, desktopFilename),
		iconPath:      filepath.Join(iconDir, iconFilename),
		installAction: true,
	}, nil
}

// Install copies the currently running executable and embedded assets to the local user
// environment directories
func Install() error {

	paths, err := newInstallPath()
	if err != nil {
		return err
	}

	showInstallStart(paths)

	// Ensure destination directories exist
	dirs := []string{paths.binDir, paths.appDir, paths.iconDir}

	for _, dir := range dirs {

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

	}

	// Copy the executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks if the binary was invoked via a symlink
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to evaluate symlinks for executable: %w", err)
	}

	if err := copyLocalFile(execPath, paths.binPath, 0755); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Copy the .desktop file
	if err := copyEmbeddedFile(desktopFilename, paths.desktopPath, 0644); err != nil {
		return fmt.Errorf("failed to copy desktop file: %w", err)
	}

	// Copy the icon file
	if err := copyEmbeddedFile(iconFilename, paths.iconPath, 0644); err != nil {
		return fmt.Errorf("failed to copy icon file: %w", err)
	}

	// Update the desktop database so the application menu/launcher refreshes
	if err := updateDesktopDatabase(paths.appDir); err != nil {
		return err
	}

	// Update the GTK icon cache so the new icon displays immediately
	updateIconCache(paths.iconDir)

	showInstallCompleted(paths)

	return nil
}

// Uninstall removes the installed BSC executable and assets from the user's local environment
func Uninstall() error {

	paths, err := newInstallPath()
	if err != nil {
		return err
	}

	paths.installAction = false
	showInstallStart(paths)

	// Remove each file, ignoring errors if the file is already gone
	filesToRemove := []string{paths.binPath, paths.desktopPath, paths.iconPath}

	for _, file := range filesToRemove {

		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}

	}

	// Update the desktop database to remove the app from the application launcher/menu
	if err := updateDesktopDatabase(paths.appDir); err != nil {
		return err
	}

	// Update the GTK icon cache so the application icon is flushed permanently
	updateIconCache(paths.iconDir)

	showInstallCompleted(paths)

	return nil
}

// updateDesktopDatabase runs the 'update-desktop-database' command on the specified application
// directory to ensure that the desktop environment immediately recognizes the changes
func updateDesktopDatabase(appDir string) error {

	cleanDir := filepath.Clean(appDir)

	// Ensure the application directory is absolute
	if !filepath.IsAbs(cleanDir) {
		return errInvalidAppDir
	}

	cmd := exec.Command("update-desktop-database", cleanDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update desktop database command failed: %w", err)
	}

	return nil
}

// updateIconCache runs 'gtk-update-icon-cache' to refresh the desktop icon cache
func updateIconCache(iconDir string) {

	// Navigate up to the base ".../icons/hicolor" directory (required)
	themeDir := filepath.Dir(filepath.Dir(iconDir))

	cmd := exec.Command("gtk-update-icon-cache", "-f", "-t", "-q", themeDir)

	// Ignore errors, as the user's system may not have GTK installed, so this shouldn't fail the
	// installation (the icon will just not update immediately)
	_ = cmd.Run()

}

// getDataHome returns the XDG_DATA_HOME directory or its standard fallback
func getDataHome(homeDir string) string {

	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" && filepath.IsAbs(dir) {
		return dir
	}

	return filepath.Join(homeDir, ".local", "share")
}

// getBinHome returns the XDG_BIN_HOME directory or its standard fallback
func getBinHome(homeDir string) string {

	if dir := os.Getenv("XDG_BIN_HOME"); dir != "" && filepath.IsAbs(dir) {
		return dir
	}

	return filepath.Join(homeDir, ".local", "bin")
}

// copyToFile copies data to a file with the specified permissions
func copyToFile(r io.Reader, dst string, perm fs.FileMode) error {

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("open destination %s: %w", dst, err)
	}

	if _, err := io.Copy(out, r); err != nil {
		out.Close()

		return fmt.Errorf("copy to %s: %w", dst, err)
	}

	// Close explicitly (not via defer) so write errors are not silently swallowed
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dst, err)
	}

	return nil
}

// copyLocalFile copies a file to a destination path with the specified permissions
func copyLocalFile(src, dst string, perm fs.FileMode) error {

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", src, err)
	}

	defer in.Close()

	return copyToFile(in, dst, perm)
}

// copyEmbeddedFile copies a file from the embedded assets to a destination path
func copyEmbeddedFile(name, dst string, perm fs.FileMode) error {

	in, err := assets.InstallerAssets.Open(name)
	if err != nil {
		return fmt.Errorf("open embedded asset %s: %w", name, err)
	}

	defer in.Close()

	return copyToFile(in, dst, perm)
}

// showInstallStart displays a pre-installation message
func showInstallStart(paths *installPaths) {

	install := "Installing"
	if !paths.installAction {
		install = "Uninstalling"
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, install+" the following "+config.GetFullVersion()+" files...")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Binary:       "+paths.binPath)
	fmt.Fprintln(os.Stdout, "Desktop file: "+paths.desktopPath)
	fmt.Fprintln(os.Stdout, "Icon:         "+paths.iconPath)
	fmt.Fprintln(os.Stdout, "")

}

// showInstallCompleted displays a post-installation message
func showInstallCompleted(paths *installPaths) {

	install := "Installation"
	if !paths.installAction {
		install = "Uninstallation"
	}

	fmt.Fprintln(os.Stdout, install+" completed successfully.")
	fmt.Fprintln(os.Stdout, "")

}
