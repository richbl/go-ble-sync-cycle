package flags

import (
	"os"
	"reflect"
	"testing"
)

const (
	TestConfigFile   = "test.toml"
	TestSeekPosition = "01:30"
)

// TestParseArgs tests the ParseArgs function
func TestParseArgs(t *testing.T) {

	testCases := getTestCases()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}

}

// getTestCases returns a slice of test cases for the ParseArgs function
func getTestCases() []struct {
	name     string
	args     []string
	wantErr  bool
	expected Flags
} {
	return []struct {
		name     string
		args     []string
		wantErr  bool
		expected Flags
	}{
		{
			name:     "no flags",
			args:     []string{},
			wantErr:  false,
			expected: Flags{Config: "", Seek: "", Help: false},
		},
		{
			name:     "all flags with long names",
			args:     []string{"--config", TestConfigFile, "--seek", TestSeekPosition, "--help"},
			wantErr:  false,
			expected: Flags{Config: TestConfigFile, Seek: TestSeekPosition, Help: true},
		},
		{
			name:     "all flags with short names",
			args:     []string{"-c", TestConfigFile, "-s", TestSeekPosition, "-h"},
			wantErr:  false,
			expected: Flags{Config: TestConfigFile, Seek: TestSeekPosition, Help: true},
		},
		{
			name:     "mixed short and long names",
			args:     []string{"-c", TestConfigFile, "--seek", TestSeekPosition, "-h"},
			wantErr:  false,
			expected: Flags{Config: TestConfigFile, Seek: TestSeekPosition, Help: true},
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid", "value"},
			wantErr: true,
		},
	}

}

// runTestCase runs a single test case for the ParseArgs function
func runTestCase(t *testing.T, tc struct {
	name     string
	args     []string
	wantErr  bool
	expected Flags
}) {
	flags = Flags{} // Reset flags

	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = append([]string{"app"}, tc.args...)

	err := ParseArgs()

	if (err != nil) != tc.wantErr {
		t.Errorf("ParseArgs() error = %v, wantErr %v", err, tc.wantErr)
		return
	}

	if !tc.wantErr && !reflect.DeepEqual(flags, tc.expected) {
		t.Errorf("ParseArgs() got = %v, want %v", flags, tc.expected)
	}

}

// TestGetFlags tests the GetFlags function
func TestGetFlags(t *testing.T) {

	// Set up test flags
	testFlags := Flags{
		Config: TestConfigFile,
		Seek:   TestSeekPosition,
		Help:   true,
	}

	// Set the package-level flags variable
	flags = testFlags

	// Test GetFlags
	result := GetFlags()

	if !reflect.DeepEqual(result, testFlags) {
		t.Errorf("GetFlags() = %v, want %v", result, testFlags)
	}

}

// TestShowHelp tests the ShowHelp function
func TestShowHelp(t *testing.T) {

	// Since ShowHelp() is a simple formatting function, we'll just test that it doesn't panic
	t.Run("ShowHelp should not panic", func(t *testing.T) {

		defer func() {

			if r := recover(); r != nil {
				t.Errorf("ShowHelp() panicked: %v", r)
			}

		}()

		ShowHelp()
	})

}

// TestFlagInfosConfiguration tests the configuration of flagInfos
func TestFlagInfosConfiguration(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		flagInfo FlagInfo
		wantType any
	}{
		{
			name:     "config flag",
			flagInfo: flagInfos[0],
			wantType: (*string)(nil),
		},
		{
			name:     "seek flag",
			flagInfo: flagInfos[1],
			wantType: (*string)(nil),
		},
		{
			name:     "help flag",
			flagInfo: flagInfos[2],
			wantType: (*bool)(nil),
		},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if reflect.TypeOf(tt.flagInfo.Result) != reflect.TypeOf(tt.wantType) {
				t.Errorf("FlagInfo %s has wrong type = %v, want %v",
					tt.flagInfo.Name,
					reflect.TypeOf(tt.flagInfo.Result),
					reflect.TypeOf(tt.wantType))
			}

		})
	}

}
