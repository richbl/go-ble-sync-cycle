package flags

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// FlagInfo holds structural information about a flag
type FlagInfo struct {
	Result    interface{} // Pointer to the resulting value
	Name      string      // Name of the flag, e.g., "config"
	ShortName string      // Short name of the flag, e.g., "c"
	Value     string      // Default value
	Usage     string      // Usage description (used for help)
}

// Flags holds a list of available command-line flags
type Flags struct {
	Config string
	Seek   string
	Help   bool
}

var (
	flags     Flags
	flagInfos = []FlagInfo{
		{
			Result:    &flags.Config,
			Name:      "config",
			ShortName: "c",
			Value:     "",
			Usage:     "Path to the configuration file ('path/to/config.toml')",
		},
		{
			Result:    &flags.Seek,
			Name:      "seek",
			ShortName: "s",
			Value:     "",
			Usage:     "Seek to a specific time in the video ('MM:SS')",
		},
		{
			Result:    &flags.Help,
			Name:      "help",
			ShortName: "h",
			Value:     "false",
			Usage:     "Display this help message",
		},
	}
)

// ParseArgs parses the command-line flags and returns an error if an undefined flag is used
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
	fmt.Println("Usage: ble-sync-cycle [flags]")
	fmt.Println("")
	fmt.Println("The following flags are available:")
	fmt.Println("")

	for _, fi := range flagInfos {
		fmt.Printf("-%s | --%s:\t%s\n", fi.ShortName, fi.Name, fi.Usage)
	}

	fmt.Println("")
}

// GetFlags returns the parsed flags
func GetFlags() Flags {
	return flags
}
