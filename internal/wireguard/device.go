package wireguard

import (
	"fmt"
	"os"
	"strings"

	"vpn-manager/internal/data"
	sshclient "vpn-manager/internal/ssh"
)

func CreateDevice(name string) (data.Device, error) {
	if err := ValidateDeviceName(name); err != nil {
		return data.Device{}, err
	}

	settings, err := data.ReadSettingsFile()
	if err != nil {
		return data.Device{}, fmt.Errorf("Read settings: %w", err)
	}

	devices, err := data.ReadDevicesFile()
	if err != nil {
		return data.Device{}, fmt.Errorf("Read devices: %w", err)
	}

	// Reject duplicate names (file names are case-insensitive on Windows)
	for _, d := range devices {
		if strings.EqualFold(d.Name, name) {
			return data.Device{}, fmt.Errorf("Device name already exists: %s", name)
		}
	}

	client, err := sshclient.ConnectServer()
	if err != nil {
		return data.Device{}, err
	}
	defer client.Close()

	// Get the server public key (only when it is not saved yet)
	if settings.ServerPublicKey == "" {
		key, err := GetServerPublicKey(client, settings.ServerInterface)
		if err != nil {
			return data.Device{}, err
		}

		settings.ServerPublicKey = key
		if err := data.WriteSettingsFile(settings); err != nil {
			return data.Device{}, fmt.Errorf("Save settings: %w", err)
		}
	}

	// IPs already used locally
	localIPs := make([]string, 0, len(devices))
	for _, d := range devices {
		localIPs = append(localIPs, d.IP)
	}

	// Read the real subnet from wg0 so it always matches the server, and fall
	// back to the saved setting when the interface cannot be queried.
	subnet := settings.VPNSubnet
	if detected, err := GetServerSubnet(client, settings.ServerInterface); err == nil {
		subnet = detected
		if detected != settings.VPNSubnet {
			settings.VPNSubnet = detected
			if err := data.WriteSettingsFile(settings); err != nil {
				return data.Device{}, fmt.Errorf("Save settings: %w", err)
			}
		}
	}

	ip, err := AllocateIP(client, settings.ServerInterface, subnet, localIPs)
	if err != nil {
		return data.Device{}, err
	}

	keys, err := GenerateKeyPair()
	if err != nil {
		return data.Device{}, err
	}

	device := data.Device{Name: name, IP: ip, PublicKey: keys.Public}

	// Render the config before touching the server, so a bad config
	// does not leave a registered peer behind
	conf, err := RenderClientConfig(BuildClientConfig(settings, device, keys.Private))
	if err != nil {
		return data.Device{}, err
	}

	if err := AddPeer(client, settings.ServerInterface, device.PublicKey, device.IP); err != nil {
		return data.Device{}, err
	}

	// Save the config file. The directory is restricted to SYSTEM and
	// Administrators because the file contains the private key.
	confDir := data.ConfDir()
	confPath := data.ConfPath(name)
	if err := os.MkdirAll(confDir, 0700); err != nil {
		_ = RemovePeer(client, settings.ServerInterface, device.PublicKey)
		return data.Device{}, fmt.Errorf("Create config directory: %w", err)
	}
	if err := data.SecureConfDir(confDir); err != nil {
		_ = RemovePeer(client, settings.ServerInterface, device.PublicKey)
		return data.Device{}, fmt.Errorf("Restrict config directory: %w", err)
	}

	if err := os.WriteFile(confPath, conf, 0600); err != nil {
		_ = RemovePeer(client, settings.ServerInterface, device.PublicKey)
		return data.Device{}, fmt.Errorf("Save client config: %w", err)
	}

	return device, nil
}
