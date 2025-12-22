package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// SessionMetadata holds the minimal information needed to display a session in the GUI
type SessionMetadata struct {
	Title    string // The session_title from the config, or filename if empty
	FilePath string // Full path to the config file
	ErrorMsg string // Error message if validation failed
	IsValid  bool   // True if the config is valid
}

// LoadSessionMetadata loads and validates a TOML config file, extracting only the session title
func LoadSessionMetadata(filePath string) (*SessionMetadata, error) {

	metadata := &SessionMetadata{
		FilePath: filePath,
		IsValid:  false,
	}

	// Attempt to load and validate the full config
	cfg, err := loadAndValidateConfig(filePath)
	if err != nil {
		metadata.ErrorMsg = err.Error()

		return nil, fmt.Errorf("failed to load session metadata from %s: %w", filePath, err)
	}

	metadata.IsValid = true

	// Extract session title or use filename as fallback
	if strings.TrimSpace(cfg.App.SessionTitle) != "" {
		metadata.Title = cfg.App.SessionTitle
	} else {
		// Use filename without extension as fallback
		filename := filepath.Base(filePath)
		metadata.Title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	return metadata, nil
}

// loadAndValidateConfig loads a config file without applying command-line flag overrides
func loadAndValidateConfig(filePath string) (*Config, error) {

	cfg := &Config{}

	// Decode the TOML file
	_, err := toml.DecodeFile(filePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidConfigFile, err)
	}

	// Validate TOML sections
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
