package config

// NewDefault returns a Config populated with default session settings, which uses videoPath as
// the video file for playback
//
// The video file is neither created nor checked here: as with any other Config, it must exist by
// the time Validate (or Load) is called
func NewDefault(videoPath string) *Config {

	return &Config{
		App: AppConfig{
			SessionTitle: "New BSC Session",
			LogLevel:     logLevelInfo,
		},
		BLE: BLEConfig{
			SensorBDAddr:    "AA:BB:CC:DD:EE:FF",
			ScanTimeoutSecs: 30,
		},
		Speed: SpeedConfig{
			WheelCircumferenceMM: 2155,
			SpeedUnits:           SpeedUnitsMPH,
			SpeedThreshold:       0.25,
			SmoothingWindow:      5,
		},
		Video: VideoConfig{
			MediaPlayer:       MediaPlayerMPV,
			FilePath:          videoPath,
			SeekToPosition:    "00:00:00",
			AutoResume:        false,
			WindowScaleFactor: 1.0,
			UpdateIntervalSec: 0.25,
			SpeedMultiplier:   0.8,
			TargetDisplayName: "",
			OnScreenDisplay: VideoOSDConfig{
				DisplayCycleSpeed:    true,
				DisplayPlaybackSpeed: true,
				DisplayTimeRemaining: true,
				FontSize:             40,
				MarginX:              20,
				MarginY:              20,
				AlignX:               alignLeft,
				AlignY:               alignTop,
				ShowOSD:              true,
			},
		},
	}

}
