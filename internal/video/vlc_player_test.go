package video

import (
	"testing"
)

// vlcPlayerFactory creates a new vlcPlayer instance for testing
func vlcPlayerFactory() (mediaPlayer, error) {
	return newVLCPlayer()
}

func TestVLCPlayerLifecycle(t *testing.T) {
	testPlayerLifecycle(t, vlcPlayerFactory)
}

func TestVLCPlayerPlaybackControls(t *testing.T) {
	testPlayerPlaybackControls(t, vlcPlayerFactory)
}

func TestVLCPlayerConfiguration(t *testing.T) {
	testPlayerConfiguration(t, vlcPlayerFactory, "VLC")
}

// TestVLCPlayerParseTimePosition tests VLC-specific time parsing logic
func TestVLCPlayerParseTimePosition(t *testing.T) {

	testCases := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{"MM:SS format", "01:30", 90000, false},
		{"SS format", "45", 45000, false},
		{"Zero seconds", "0", 0, false},
		{"Invalid MM:SS", "01:60", 0, true},
		{"Invalid format", "abc", 0, true},
		{"Negative seconds", "-10", 0, true},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTimePosition(tc.input)

			if (err != nil) != tc.hasError {
				t.Fatalf("parseTimePosition() error = %v, wantErr %v", err, tc.hasError)
			}

			if !tc.hasError && got != tc.expected {
				t.Errorf("parseTimePosition() = %v, want %v", got, tc.expected)
			}

		})
	}
}
