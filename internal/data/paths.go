package data

import (
	"os"
	"path/filepath"
)

// appDirName is the folder under %ProgramData% that holds all runtime data.
const appDirName = "VPNManager"

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
