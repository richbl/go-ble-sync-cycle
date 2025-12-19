package flags

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// ModeType represents the current mode of operation for the application
type ModeType int

const (
	CLI ModeType = iota
	GUI
)

// FlagInfo holds structural information about a flag
type FlagInfo struct {
	Result    any      // Pointer to the resulting value
	Name      string   // Name of the flag, e.g., "config"
	ShortName string   // Short name of the flag, e.g., "c"
	Value     string   // Default value
	Usage     string   // Usage description (used for help)
	Mode      ModeType // Mode of operation (CLI or GUI)
}

// CLIFlags holds a list of available command-line flags
type CLIFlags struct {
	Logging bool
	NoGUI   bool
	Config  string
	Seek    string
	Help    bool
}

var (
	flags CLIFlags

	flagInfos = []FlagInfo{
		{
			Result:    &flags.Logging,
			Name:      "log-console",
			ShortName: "l",
			Value:     "false",
			Usage:     "Enable logging to the console",
			Mode:      GUI,
		},
		{
			Result:    &flags.NoGUI,
			Name:      "no-gui",
			ShortName: "n",
			Value:     "false",
			Usage:     "Run the application without a graphical user interface (GUI)",
			Mode:      CLI,
		},
		{
			Result:    &flags.Config,
			Name:      "config",
			ShortName: "c",
			Value:     "",
			Usage:     "Path to the configuration file ('path/to/config.toml')",
			Mode:      CLI,
		},
		{
			Result:    &flags.Seek,
			Name:      "seek",
			ShortName: "s",
			Value:     "",
			Usage:     "Seek to a specific time in the video ('MM:SS')",
			Mode:      CLI,
		},
		{
			Result:    &flags.Help,
			Name:      "help",
			ShortName: "h",
			Value:     "false",
			Usage:     "Display this help message",
			Mode:      CLI,
		},
	}
)

// ParseArgs parses the command-line flags and returns an error if an undefined flag is found
func ParseArgs() error {

	// Create a custom FlagSet
	fs := flag.NewFlagSet("app", flag.ContinueOnError)

	fs.SetOutput(io.Discard) // Suppress all output (important)

	// Register all flags
	for _, fi := range flagInfos {

		switch v := fi.Result.(type) {

		case *string:
			fs.StringVar(v, fi.Name, fi.Value, fi.Usage)
			fs.StringVar(v, fi.ShortName, fi.Value, fi.Usage)

		case *bool:
			fs.BoolVar(v, fi.Name, fi.Value == "true", fi.Usage)
			fs.BoolVar(v, fi.ShortName, fi.Value == "true", fi.Usage)
		}

	}

	// Parse the flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil
}

// ShowHelp displays application help information
func ShowHelp() {

	fmt.Println("")
	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Println("")
	fmt.Println("Usage: ble-sync-cycle [flags]")
	fmt.Println("")
	fmt.Println("The following flags are available when running in console/CLI mode:")
	fmt.Println("")

	for _, fi := range flagInfos {
		if fi.Mode == CLI {
			fmt.Printf("  -%s, --%-12s %s\n", fi.ShortName, fi.Name, fi.Usage)
		}
	}

	fmt.Println("")
	fmt.Println("The following flags are available when running in GUI mode:")
	fmt.Println("")

	for _, fi := range flagInfos {
		if fi.Mode == GUI {
			fmt.Printf("  -%s, --%-12s %s\n", fi.ShortName, fi.Name, fi.Usage)
		}
	}
	fmt.Println("")
	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Println("")
}

// Flags returns the parsed flags
func Flags() CLIFlags {
	return flags
}

// IsCLIMode checks if the user provided the flag to run in CLI-only mode
func IsCLIMode() bool {
	return flags.NoGUI
}

// IsHelpFlag checks if the user provided the flag to display help
func IsHelpFlag() bool {
	return flags.Help
}

// IsGUIConsoleLogging returns true/false to enable CLI logging while running in GUI mode
func IsGUIConsoleLogging() bool {
	return flags.Logging
}
