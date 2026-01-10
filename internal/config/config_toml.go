package config

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// commentColumn is the column position where comments start in the TOML file
const commentColumn = 40

// ConfigTemplate defines the TOML structure with inline comments
const ConfigTemplate = `# BLE Sync Cycle Configuration (TOML)
# {{.Version}}

[app]
  session_title = "{{.App.SessionTitle}}"{{pad (printf "session_title = \"%s\"" .App.SessionTitle)}}# Short description of the current cycling session (0-200 characters)
  logging_level = "{{.App.LogLevel}}"{{pad (printf "logging_level = \"%s\"" .App.LogLevel)}}# Log messages generated during execution ("debug", "info", "warn", "error")

[ble]
  sensor_bd_addr = "{{.BLE.SensorBDAddr}}"{{pad (printf "sensor_bd_addr = \"%s\"" .BLE.SensorBDAddr)}}# The Bluetooth Device Address (BD_ADDR) of the BLE peripheral
  scan_timeout_secs = {{.BLE.ScanTimeoutSecs}}{{pad (printf "scan_timeout_secs = %d" .BLE.ScanTimeoutSecs)}}# Time to wait for a response from the peripheral before connect fails (1-100 seconds)

[speed]
  wheel_circumference_mm = {{.Speed.WheelCircumferenceMM}}{{pad (printf "wheel_circumference_mm = %d" .Speed.WheelCircumferenceMM)}}# Wheel circumference (50-3000 millimeters)
  speed_units = "{{.Speed.SpeedUnits}}"{{pad (printf "speed_units = \"%s\"" .Speed.SpeedUnits)}}# The unit of measurement for speed ("mph" or "km/h")
  speed_threshold = {{printf "%.2f" .Speed.SpeedThreshold}}{{pad (printf "speed_threshold = %.2f" .Speed.SpeedThreshold)}}# Minimum speed change to trigger video playback update (0.00-10.00)
  smoothing_window = {{.Speed.SmoothingWindow}}{{pad (printf "smoothing_window = %d" .Speed.SmoothingWindow)}}# Number of recent speed readings to generate a stable moving average (1-25)

[video]
  media_player = "{{.Video.MediaPlayer}}"{{pad (printf "media_player = \"%s\"" .Video.MediaPlayer)}}# The video playback back-end to use ("mpv" or "vlc")
  file_path = "{{.Video.FilePath}}"{{pad (printf "file_path = \"%s\"" .Video.FilePath)}}# File path to the video file for playback
  seek_to_position = "{{.Video.SeekToPosition}}"{{pad (printf "seek_to_position = \"%s\"" .Video.SeekToPosition)}}# Starting playback position in the video ("MM:SS")
  window_scale_factor = {{printf "%.1f" .Video.WindowScaleFactor}}{{pad (printf "window_scale_factor = %.1f" .Video.WindowScaleFactor)}}# Scales the size of the video window (0.1-1.0, where 1.0 = full screen)
  update_interval_secs = {{printf "%.1f" .Video.UpdateIntervalSec}}{{pad (printf "update_interval_secs = %.1f" .Video.UpdateIntervalSec)}}# Frequency that the video player is sent speed updates (0.10-3.00 seconds)
  speed_multiplier = {{printf "%.1f" .Video.SpeedMultiplier}}{{pad (printf "speed_multiplier = %.1f" .Video.SpeedMultiplier)}}# Multiplier to control video playback rate (0.1-1.5, where 0.1 = slower, 1.0 = normal, 1.5 = faster playback)

[video.OSD]
  display_cycle_speed = {{.Video.OnScreenDisplay.DisplayCycleSpeed}}{{pad (printf "display_cycle_speed = %t" .Video.OnScreenDisplay.DisplayCycleSpeed)}}# Display the current cycle speed on the on-screen display (true/false)
  display_playback_speed = {{.Video.OnScreenDisplay.DisplayPlaybackSpeed}}{{pad (printf "display_playback_speed = %t" .Video.OnScreenDisplay.DisplayPlaybackSpeed)}}# Display the current video playback speed on the on-screen display (true/false)
  display_time_remaining = {{.Video.OnScreenDisplay.DisplayTimeRemaining}}{{pad (printf "display_time_remaining = %t" .Video.OnScreenDisplay.DisplayTimeRemaining)}}# Display the current video time remaining on the on-screen display (true/false)
  font_size = {{.Video.OnScreenDisplay.FontSize}}{{pad (printf "font_size = %d" .Video.OnScreenDisplay.FontSize)}}# Font size of the on-screen display (10-200 pixels)
  margin_left = {{.Video.OnScreenDisplay.MarginX}}{{pad (printf "margin_left = %d" .Video.OnScreenDisplay.MarginX)}}# Offset of the OSD from the left of the media player window (0-100 pixels)
  margin_top = {{.Video.OnScreenDisplay.MarginY}}{{pad (printf "margin_top = %d" .Video.OnScreenDisplay.MarginY)}}# Offset of the OSD from the top of the media player window (0-100 pixels)
`

// tomlContent wraps Config with version info for TOML template creation
type tomlContent struct {
	*Config
	Version string
}

// Save writes the TOML configuration to file with inline comments
func Save(filePath string, cfg *Config, version string) error {

	// Create template with custom function
	tmpl := template.New("config").Funcs(template.FuncMap{
		"pad": padToColumn,
	})

	// Parse the template
	tmpl, err := tmpl.Parse(ConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// Create a new file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	defer f.Close()

	// Create the template data
	templateData := tomlContent{
		Config:  cfg,
		Version: "v" + version,
	}

	// Merge the data with the template
	if err := tmpl.Execute(f, templateData); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// padToColumn calculates padding needed to align comments at commentColumn
func padToColumn(kvPair string) string {

	currentLen := len(kvPair)
	if currentLen >= commentColumn {
		return "  " // Minimum 2 spaces if line is too long
	}

	spacesNeeded := commentColumn - currentLen

	return strings.Repeat(" ", spacesNeeded)
}
