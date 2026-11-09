package data

import (
	"os"
	"path/filepath"
)

// appDirName is the folder under %ProgramData% that holds all runtime data.
const appDirName = "VPNManager"

// Legacy relative paths used before data moved to %ProgramData%.
const (
	legacySettingsFile = "internal/data/json/settings.json"
	legacyDevicesFile  = "internal/data/json/devices.json"
)

// AppDir returns the absolute path to the application data directory.
func AppDir() string {
	programData := os.Getenv("ProgramData")
	if programData == "" {
		drive := os.Getenv("SystemDrive")
		if drive == "" {
			drive = "C:"
		}
		programData = filepath.Join(drive, "ProgramData")
	}
	return filepath.Join(programData, appDirName)
}

// SettingsPath returns the absolute path to settings.json.
func SettingsPath() string {
	return filepath.Join(AppDir(), "settings.json")
}

// DevicesPath returns the absolute path to devices.json.
func DevicesPath() string {
	return filepath.Join(AppDir(), "devices.json")
}

// ConfDir returns the absolute path to the directory that holds client configs.
func ConfDir() string {
	return filepath.Join(AppDir(), "conf")
}

// ConfPath returns the absolute path to the config file for a device.
func ConfPath(name string) string {
	return filepath.Join(ConfDir(), name+".conf")
}

// LogPath returns the absolute path to the application log file.
func LogPath() string {
	return filepath.Join(AppDir(), "logs", "vpn-manager.log")
}

// migrateLegacyFile copies a legacy data file to its new location once.
func migrateLegacyFile(legacy, dst string) bool {
	data, err := os.ReadFile(legacy)
	if err != nil {
		return false
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return false
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return false
	}

	return true
}
