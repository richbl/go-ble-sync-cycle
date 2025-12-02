package video

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	config "github.com/richbl/go-ble-sync-cycle/internal/config"
	logger "github.com/richbl/go-ble-sync-cycle/internal/logger"
	speed "github.com/richbl/go-ble-sync-cycle/internal/speed"
)

// mockMediaPlayer is a mock implementation of the mediaPlayer interface for testing
type mockMediaPlayer struct {
	mu                  sync.Mutex
	calls               map[string]int
	lastShowText        string
	lastSpeed           float64
	lastPauseState      bool
	loadFileErr         error
	setupEventsErr      error
	setFullscreenErr    error
	setKeepOpenErr      error
	SetOSDErr           error
	seekErr             error
	setSpeedErr         error
	setPauseErr         error
	showTextErr         error
	getRemainingTime    int64
	getRemainingTimeErr error
	eventChan           chan *playerEvent
}

// updateSpeedTestCase defines a test case for updateSpeedFromController
type updateSpeedTestCase struct {
	name              string
	currentSpeed      float64
	lastSpeed         float64
	speedThreshold    float64
	expectPause       bool
	expectedPauseCall bool
	expectedSpeedCall bool
}

// testConfigData holds test constants and configurations
var testConfigData = struct {
	filename        string
	windowScale     float64
	SeekToPosition  string
	updateInterval  float64
	speedMultiplier float64
	speedThreshold  float64
}{
	filename:        "test_video.mp4",
	windowScale:     1.0,
	SeekToPosition:  "0.25",
	updateInterval:  1.0,
	speedMultiplier: 1.0,
	speedThreshold:  0.1,
}

// init is called to set the log level for tests
func init() {
	logger.Initialize("debug")
}

// createTestConfig returns test video and speed configurations
func createTestConfig() (config.VideoConfig, config.SpeedConfig) {

	vc := config.VideoConfig{
		MediaPlayer:       config.MediaPlayerMPV,
		FilePath:          testConfigData.filename,
		WindowScaleFactor: testConfigData.windowScale,
		SeekToPosition:    testConfigData.SeekToPosition,
		UpdateIntervalSec: testConfigData.updateInterval,
		SpeedMultiplier:   testConfigData.speedMultiplier,
		OnScreenDisplay: config.VideoOSDConfig{
			FontSize:             24,
			DisplayCycleSpeed:    true,
			DisplayPlaybackSpeed: true,
			DisplayTimeRemaining: true,
			ShowOSD:              true,
		},
	}

	sc := config.SpeedConfig{
		SpeedThreshold: testConfigData.speedThreshold,
		SpeedUnits:     config.SpeedUnitsMPH,
	}

	return vc, sc
}

// newMockMediaPlayer creates a new mockMediaPlayer instance
func newMockMediaPlayer() *mockMediaPlayer {

	return &mockMediaPlayer{
		calls:     make(map[string]int),
		eventChan: make(chan *playerEvent, 1),
	}

}

// recordCall records a method call by name
func (m *mockMediaPlayer) recordCall(name string) {

	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[name]++

}

// callCount returns the number of times a method was called
func (m *mockMediaPlayer) callCount(name string) int {

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calls[name]
}

// waitEvent waits for a player event or times out
func (m *mockMediaPlayer) terminatePlayer() {
	m.recordCall("terminatePlayer")
}

// loadFile loads a video file into the mock media player
func (m *mockMediaPlayer) loadFile(_ string) error {

	m.recordCall("loadFile")
	return m.loadFileErr

}

// setupEvents sets up event handling for the mock media player
func (m *mockMediaPlayer) setupEvents() error {

	m.recordCall("setupEvents")
	return m.setupEventsErr

}

// seek seeks to a specified position in the video
func (m *mockMediaPlayer) setFullscreen(_ bool) error {

	m.recordCall("setFullscreen")
	return m.setFullscreenErr

}

// setKeepOpen configures whether the player window stays open after playback
func (m *mockMediaPlayer) setKeepOpen(_ bool) error {

	m.recordCall("setKeepOpen")
	return m.setKeepOpenErr

}

// setOSD configures the On-Screen Display (OSD)
func (m *mockMediaPlayer) setOSD(_ osdConfig) error {

	m.recordCall("setOSD")
	return m.SetOSDErr

}

// seek seeks to a specified position in the video
func (m *mockMediaPlayer) seek(_ string) error {

	m.recordCall("seek")
	return m.seekErr

}

// setSpeed sets the playback speed of the video
func (m *mockMediaPlayer) setSpeed(speed float64) error {

	m.recordCall("setSpeed")
	m.lastSpeed = speed

	return m.setSpeedErr
}

// setPause sets the pause state of the video
func (m *mockMediaPlayer) setPause(pause bool) error {

	m.recordCall("setPause")
	m.lastPauseState = pause

	return m.setPauseErr
}

// showOSDText displays text on the OSD
func (m *mockMediaPlayer) showOSDText(text string) error {

	m.recordCall("showOSDText")
	m.lastShowText = text

	return m.showTextErr
}

// getTimeRemaining gets the remaining time of the video
func (m *mockMediaPlayer) getTimeRemaining() (int64, error) {

	m.recordCall("getTimeRemaining")
	return m.getRemainingTime, m.getRemainingTimeErr

}

// waitEvent waits for a player event or times out
func (m *mockMediaPlayer) waitEvent(timeout float64) *playerEvent {

	m.recordCall("waitEvent")

	select {
	case e := <-m.eventChan:
		return e
	case <-time.After(time.Duration(timeout * float64(time.Second))):
		return &playerEvent{id: eventNone}
	}

}

// TestNewPlaybackController tests the NewPlaybackController function
func TestNewPlaybackController(t *testing.T) {

	vc, sc := createTestConfig()

	t.Run("successful creation", func(t *testing.T) {
		controller, err := NewPlaybackController(vc, sc)
		if err != nil {
			t.Skip("Skipping test: cannot create real player in unit test environment.")
		}

		if controller == nil {
			t.Error("controller should not be nil")
		}

	})

	t.Run("unsupported media player", func(t *testing.T) {
		vcInvalid := vc
		vcInvalid.MediaPlayer = "invalid-player"
		controller, err := NewPlaybackController(vcInvalid, sc)
		if err == nil {
			t.Error("expected an error for unsupported media player, but got nil")
		}

		if controller != nil {
			t.Error("controller should be nil on error")
		}

	})
}

// setupTestController creates a PlaybackController with a mock player for testing
func setupTestController(t *testing.T) (*PlaybackController, *mockMediaPlayer, *speed.Controller) {

	t.Helper()
	vc, sc := createTestConfig()
	mockPlayer := newMockMediaPlayer()
	speedCtrl := speed.NewSpeedController(5)
	controller := &PlaybackController{
		videoConfig: vc,
		speedConfig: sc,
		osdConfig: osdConfig{
			showOSD:              vc.OnScreenDisplay.ShowOSD,
			fontSize:             vc.OnScreenDisplay.FontSize,
			displayCycleSpeed:    vc.OnScreenDisplay.DisplayCycleSpeed,
			displayPlaybackSpeed: vc.OnScreenDisplay.DisplayPlaybackSpeed,
			displayTimeRemaining: vc.OnScreenDisplay.DisplayTimeRemaining,
		},
		player:     mockPlayer,
		speedState: &speedState{},
	}

	return controller, mockPlayer, speedCtrl
}

// TestPlaybackController_Start tests the Start method of PlaybackController
func TestPlaybackControllerStart(t *testing.T) {

	controller, mockPlayer, speedCtrl := setupTestController(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var startErr error
	var wg sync.WaitGroup

	// Start the controller in a separate goroutine
	wg.Go(func() {
		startErr = controller.Start(ctx, speedCtrl)
	})

	// Allow time for initial setup
	time.Sleep(100 * time.Millisecond)

	t.Run("initialization calls", func(t *testing.T) {
		verifyInitializationCalls(t, mockPlayer)
	})

	// Cancel context to stop the loop and trigger cleanup
	cancel()
	wg.Wait()

	t.Run("termination and error check", func(t *testing.T) {
		verifyTermination(t, mockPlayer, startErr)
	})
}

// verifyInitializationCalls checks that the expected player methods are called during startup
func verifyInitializationCalls(t *testing.T, mockPlayer *mockMediaPlayer) {

	t.Helper()
	expectedCalls := map[string]int{
		"loadFile":      1,
		"setupEvents":   1,
		"setFullscreen": 1,
		"setKeepOpen":   1,
		"setOSD":        1,
		"seek":          1,
	}

	for method, count := range expectedCalls {

		if mockPlayer.callCount(method) != count {
			t.Errorf("expected %s to be called %d time(s), got %d", method, count, mockPlayer.callCount(method))
		}

	}
}

// verifyTermination checks for clean shutdown and no errors
func verifyTermination(t *testing.T, mockPlayer *mockMediaPlayer, startErr error) {

	t.Helper()

	if startErr != nil {
		t.Errorf("Start() returned an unexpected error: %v", startErr)
	}

	if mockPlayer.callCount("terminatePlayer") != 1 {
		t.Errorf("expected terminatePlayer to be called once, got %d", mockPlayer.callCount("terminatePlayer"))
	}

}

// updateSpeedTestCase defines a test case for updateSpeedFromController
func runSingleUpdateSpeedTest(t *testing.T, vc config.VideoConfig, sc config.SpeedConfig, tc updateSpeedTestCase) {

	mockPlayer := newMockMediaPlayer()
	localSC := sc
	localSC.SpeedThreshold = tc.speedThreshold

	controller := &PlaybackController{
		videoConfig: vc,
		speedConfig: localSC,
		osdConfig:   osdConfig{showOSD: true},
		player:      mockPlayer,
		speedState:  &speedState{last: tc.lastSpeed},
	}
	controller.speedUnitMultiplier = 0.1 // For simplicity

	// Create a fresh speed controller per test to avoid cross-test state
	speedCtrl := speed.NewSpeedController(5)

	// Fill the speed controller's buffer to get a predictable smoothed speed
	for range 5 {
		speedCtrl.UpdateSpeed(tc.currentSpeed)
	}

	err := controller.updateSpeedFromController(speedCtrl)
	if err != nil {
		t.Fatalf("updateSpeedFromController() returned an error: %v", err)
	}

	// Verify setPause calls and state
	pauseCalls := mockPlayer.callCount("setPause")

	if tc.expectedPauseCall {

		if pauseCalls != 1 {
			t.Errorf("expected setPause to be called once, got %d", pauseCalls)
		}

		if mockPlayer.lastPauseState != tc.expectPause {
			t.Errorf("expected pause state to be %v, got %v", tc.expectPause, mockPlayer.lastPauseState)
		}

	} else if pauseCalls > 0 {
		t.Errorf("expected setPause not to be called, but was called %d times", pauseCalls)
	}

	// Verify setSpeed calls
	speedCalls := mockPlayer.callCount("setSpeed")

	if tc.expectedSpeedCall {

		if speedCalls != 1 {
			t.Errorf("expected setSpeed to be called once, got %d", speedCalls)
		}
	} else if speedCalls > 0 {
		t.Errorf("expected setSpeed not to be called, but was called %d times", speedCalls)
	}

}

// TestUpdateSpeedFromController tests the updateSpeedFromController method
func TestUpdateSpeedFromController(t *testing.T) {

	vc, sc := createTestConfig()

	testCases := []updateSpeedTestCase{
		{
			name:              "zero speed",
			currentSpeed:      0.0,
			lastSpeed:         10.0,
			speedThreshold:    0.2, // this value doesn't matter for this test, but good to be explicit
			expectPause:       true,
			expectedPauseCall: true,
			expectedSpeedCall: false,
		},
		{
			name:              "speed below threshold",
			currentSpeed:      10.1,
			lastSpeed:         10.0,
			speedThreshold:    0.2,
			expectPause:       false,
			expectedPauseCall: false,
			expectedSpeedCall: false,
		},
		{
			name:              "speed above threshold",
			currentSpeed:      10.3,
			lastSpeed:         10.0,
			speedThreshold:    0.2,
			expectPause:       false,
			expectedPauseCall: true,
			expectedSpeedCall: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runSingleUpdateSpeedTest(t, vc, sc, tc)
		})
	}

}

// TestHandleZeroSpeed tests the handleZeroSpeed method
func TestFormatSeconds(t *testing.T) {

	testCases := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{"zero seconds", 0, "00:00:00"},
		{"less than a minute", 59, "00:00:59"},
		{"one minute", 60, "00:01:00"},
		{"one hour", 3600, "01:00:00"},
		{"complex time", 3661, "01:01:01"},
		{"large time", 86399, "23:59:59"}, // 24 hours - 1 second
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if got := formatSeconds(tc.seconds); got != tc.expected {
				t.Errorf("formatSeconds(%d) = %q, want %q", tc.seconds, got, tc.expected)
			}

		})
	}

}

// TestUpdateDisplay tests the updateDisplay method of PlaybackController
func TestUpdateDisplay(t *testing.T) {

	vc, sc := createTestConfig()
	mockPlayer := newMockMediaPlayer()

	controller := &PlaybackController{
		videoConfig: vc,
		speedConfig: sc,
		osdConfig: osdConfig{
			showOSD:              true,
			displayCycleSpeed:    true,
			displayPlaybackSpeed: true,
			displayTimeRemaining: true,
		},
		player:     mockPlayer,
		speedState: &speedState{},
	}

	t.Run("paused display", func(t *testing.T) {

		err := controller.updateDisplay(0.0, 0.0)
		if err != nil {
			t.Fatalf("updateDisplay failed: %v", err)
		}

		if mockPlayer.lastShowText != "Paused" {
			t.Errorf("expected OSD text 'Paused', got %q", mockPlayer.lastShowText)
		}

	})

	t.Run("active display", func(t *testing.T) {

		mockPlayer.getRemainingTime = 125 // 00:02:05
		err := controller.updateDisplay(15.5, 1.55)
		if err != nil {
			t.Fatalf("updateDisplay failed: %v", err)
		}

		var expectedText bytes.Buffer
		var errFailedToWriteExpectedText = "failed to write expected text: %v"

		if _, err := expectedText.WriteString("Cycle Speed: 15.5 mph\n"); err != nil {
			t.Fatalf(errFailedToWriteExpectedText, err)
		}

		if _, err := expectedText.WriteString("Playback Speed: 1.55x\n"); err != nil {
			t.Fatalf(errFailedToWriteExpectedText, err)
		}

		if _, err := expectedText.WriteString("Time Remaining: 00:02:05\n"); err != nil {
			t.Fatalf(errFailedToWriteExpectedText, err)
		}

		if mockPlayer.lastShowText != expectedText.String() {
			t.Errorf("unexpected OSD text\ngot:  %q\nwant: %q", mockPlayer.lastShowText, expectedText.String())
		}
	})
}
