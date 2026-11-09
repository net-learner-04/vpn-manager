package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Default VPN subnet used when the setting is missing or still the old value
const defaultVPNSubnet = "10.10.10.0/24"

type Settings struct {
	// Server Connection
	ServerAddress     string `json:"server_address"`
	SSHPort           string `json:"ssh_port"`
	SSHUsername       string `json:"ssh_username"`
	SSHKeyPath        string `json:"ssh_key_path"`
	SSHKnownHostsPath string `json:"known_hosts_path"`
	ConnectionTimeout string `json:"connection_timeout"`

	// WireGuard
	ServerInterface string `json:"server_interface"`
	WireGuardPort   string `json:"wireguard_port"`
	DNS             string `json:"dns"`
	AllowedIPs      string `json:"allowed_ips"`
	ServerPublicKey string `json:"server_public_key"`
	RefreshInterval string `json:"refresh_interval"`
	VPNSubnet       string `json:"vpn_subnet"`
}

func CreateSettingsFile() error {
	path := SettingsPath()

	_, err := os.Stat(path)
	if err == nil {
		// The file already exists
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		// Errors Other Than “File Not Found”
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// Default SSH Key and Known Hosts Paths
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Default Json Formats in Settings
	config := Settings{
		// Server Connection
		ServerAddress:     "",
		SSHPort:           "22",
		SSHUsername:       "",
		SSHKeyPath:        "",
		SSHKnownHostsPath: knownHostsPath,
		ConnectionTimeout: "10",

		// WireGuard
		ServerInterface: "wg0",
		WireGuardPort:   "51820",
		DNS:             "1.1.1.1",
		AllowedIPs:      "0.0.0.0/0",
		RefreshInterval: "5",
		VPNSubnet:       defaultVPNSubnet,
	}

	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	// Create the parent directory if it does not exist yet
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// If the file doesn't exist, create it
	if err := os.WriteFile(path, data, 0644); err != nil {
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
	data, err := os.ReadFile(SettingsPath())
	if err != nil {
		return config, err
	}

	// Convert JSON Data to a Settings Structure
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	// Migrate a missing or stale VPN subnet so device creation always has a
	// subnet that matches the server.
	if config.VPNSubnet == "" || config.VPNSubnet == "10.0.0.0/24" {
		config.VPNSubnet = defaultVPNSubnet
		if err := WriteSettingsFile(config); err != nil {
			return config, err
		}
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
	err = os.WriteFile(SettingsPath(), jsondata, 0644)
	if err != nil {
		return err
	}

	return nil
}
