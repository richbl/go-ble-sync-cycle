package config

import (
	"fmt"
	"regexp"
	"strings"
)

// BLEConfig defines Bluetooth Low Energy settings from the TOML config file
type BLEConfig struct {
	SensorBDAddr    string `toml:"sensor_bd_addr"`
	ScanTimeoutSecs int    `toml:"scan_timeout_secs"`
}

// validate checks BLEConfig for valid settings
func (bc *BLEConfig) validate() error {

	if err := validateField(bc.ScanTimeoutSecs, 1, 100, errInvalidScanTimeout); err != nil {
		return err
	}

	pattern := `^([0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5})$`
	re := regexp.MustCompile(pattern)

	if !re.MatchString(strings.TrimSpace(bc.SensorBDAddr)) {
		return fmt.Errorf(errFormatRev, errInvalidBDAddr, bc.SensorBDAddr)
	}

	return nil
}
