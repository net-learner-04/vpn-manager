package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// Specify the file path where you want to save it later
const settingsfile = "internal/data/settings.json"

type Settings struct {
	// Server Connection
	ServerAddress     string `json:"server_address"`
	SSHPort           string `json:"ssh_port"`
	SSHUsername       string `json:"ssh_username"`
	SSHKeyPath        string `json:"ssh_key_path"`
	ConnectionTimeout string `json:"connection_timeout"`

	// WireGuard
	WireGuardInterface string `json:"wireguard_interface"`
	RefreshInterval    string `json:"refresh_interval"`

	// Appearance & Behavior
	CloseToTray    bool `json:"close_to_tray"`
	AutoRefresh    bool `json:"auto_refresh"`
	StartMinimized bool `json:"start_minimized"`
}

func CreateSettingsFile() error {
	_, err := os.Stat(settingsfile)
	if err == nil {
		// The file already exists
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		// Errors Other Than “File Not Found”
		return err
	}

	// Default Json Formats in Settings
	config := Settings{
		// Server Connection
		ServerAddress:     "",
		SSHPort:           "22",
		SSHUsername:       "",
		SSHKeyPath:        "",
		ConnectionTimeout: "10",

		// WireGuard
		WireGuardInterface: "wg0",
		RefreshInterval:    "5",

		// Appearance & Behavior
		CloseToTray:    false,
		AutoRefresh:    true,
		StartMinimized: false,
	}

	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	// If the file doesn't exist, create it
	if err := os.WriteFile(settingsfile, data, 0644); err != nil {
		return err
	}

	return nil
}

func ReadSettingsFile() (Settings, error) {
	var config Settings

	// Confirmation of Configuration File Creation
	if err := CreateSettingsFile(); err != nil {
		return config, err
	}

	// Reading a JSON File
	data, err := os.ReadFile(settingsfile)
	if err != nil {
		return config, err
	}

	// Convert JSON Data to a Settings Structure
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}

func WriteSettingsFile(config Settings) error {
	// Convert Settings to JSON
	jsondata, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	// Save JSON file
	err = os.WriteFile(settingsfile, jsondata, 0644)
	if err != nil {
		return err
	}

	return nil
}
