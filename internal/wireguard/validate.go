package wireguard

import (
	"errors"
	"fmt"
	"strings"
)

// maxDeviceNameLength is the longest allowed device name.
const maxDeviceNameLength = 32

// ValidateDeviceName checks that a device name is safe to use as a config
// file name and a Windows service name.
func ValidateDeviceName(name string) error {
	if name == "" {
		return errors.New("Device name cannot be empty")
	}

	if len(name) > maxDeviceNameLength {
		return fmt.Errorf("Device name is too long (max %d characters)", maxDeviceNameLength)
	}

	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_'
		if !ok {
			return errors.New("Device name can only contain letters, numbers, '-' and '_'")
		}
	}

	// The name becomes a file name, so reject Windows reserved names.
	base := strings.ToLower(name)
	if i := strings.IndexByte(base, '.'); i != -1 {
		base = base[:i]
	}
	switch base {
	case "con", "prn", "aux", "nul",
		"com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9",
		"lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9":
		return errors.New("Device name is a reserved Windows name")
	}

	return nil
}
