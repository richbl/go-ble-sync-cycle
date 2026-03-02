package config

import (
	"fmt"
	"os"
)

// VideoConfig defines video playback and display settings from the TOML config file
type VideoConfig struct {
	MediaPlayer       string         `toml:"media_player"`
	FilePath          string         `toml:"file_path"`
	SeekToPosition    string         `toml:"seek_to_position"`
	WindowScaleFactor float64        `toml:"window_scale_factor"`
	UpdateIntervalSec float64        `toml:"update_interval_secs"`
	SpeedMultiplier   float64        `toml:"speed_multiplier"`
	OnScreenDisplay   VideoOSDConfig `toml:"OSD"`
}

// VideoOSDConfig defines on-screen display settings for video playback from the TOML config file
type VideoOSDConfig struct {
	FontSize             int    `toml:"font_size"`
	MarginX              int    `toml:"margin_x"`
	MarginY              int    `toml:"margin_y"`
	AlignX               string `toml:"align_x"`
	AlignY               string `toml:"align_y"`
	DisplayCycleSpeed    bool   `toml:"display_cycle_speed"`
	DisplayPlaybackSpeed bool   `toml:"display_playback_speed"`
	DisplayTimeRemaining bool   `toml:"display_time_remaining"`
	ShowOSD              bool   `toml:"-"`
}

// validate checks VideoConfig for valid settings
func (vc *VideoConfig) validate() error {

	if err := checkForVideoFile(vc.FilePath); err != nil {
		return err
	}

	validPlayer := map[string]bool{
		MediaPlayerMPV: true,
	}

	validAlignX := map[string]bool{
		"left":   true,
		"center": true,
		"right":  true,
	}

	validAlignY := map[string]bool{
		"top":    true,
		"center": true,
		"bottom": true,
	}

	if !validPlayer[vc.MediaPlayer] {
		return fmt.Errorf(errFormatRev, errInvalidPlayer, vc.MediaPlayer)
	}

	if !validAlignX[vc.OnScreenDisplay.AlignX] {
		return fmt.Errorf(errFormatRev, errInvalidAlignX, vc.OnScreenDisplay.AlignX)
	}

	if !validAlignY[vc.OnScreenDisplay.AlignY] {
		return fmt.Errorf(errFormatRev, errInvalidAlignY, vc.OnScreenDisplay.AlignY)
	}

	if err := validateConfigFields(vc.configValidationRanges()); err != nil {
		return err
	}

	if !validateTimeFormat(vc.SeekToPosition) {
		return fmt.Errorf(errFormatRev, errInvalidSeek, vc.SeekToPosition)
	}

	// Compute ShowOSD state based on display settings in TOML config file
	vc.OnScreenDisplay.ShowOSD = vc.OnScreenDisplay.DisplayCycleSpeed ||
		vc.OnScreenDisplay.DisplayPlaybackSpeed || vc.OnScreenDisplay.DisplayTimeRemaining

	return nil
}

// configValidationRanges returns validation ranges for VideoConfig
func (vc *VideoConfig) configValidationRanges() *[]validationRange {

	return &[]validationRange{
		{vc.WindowScaleFactor, 0.1, 1.0, errWindowScale},
		{vc.UpdateIntervalSec, 0.1, 3.0, errInvalidInterval},
		{vc.SpeedMultiplier, 0.1, 1.5, errSpeedMultiplier},
		{vc.OnScreenDisplay.FontSize, 10, 200, errFontSize},
		{vc.OnScreenDisplay.MarginX, 0, 300, errOSDMargin},
		{vc.OnScreenDisplay.MarginY, 0, 600, errOSDMargin},
	}

}

// checkForVideoFile checks if the provided file exists
func checkForVideoFile(filename string) error {

	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf(errFormat, errVideoFile, err)
	}

	return nil
}
