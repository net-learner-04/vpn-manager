package wireguard

import (
	"fmt"
	"net"
	"regexp"

	"vpn-manager/internal/data"
	sshclient "vpn-manager/internal/ssh"

	"golang.org/x/crypto/ssh"
)

// usernamePattern matches a safe SSH username for the sudoers file.
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// SetupServer installs WireGuard on the server and creates the wg0 interface
// when it does not exist yet.
func SetupServer() error {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return fmt.Errorf("Read settings: %w", err)
	}

	client, err := sshclient.ConnectServer()
	if err != nil {
		return err
	}
	defer client.Close()

	// Install WireGuard when it is missing
	if _, err := sshclient.RunCommand(client, "command -v wg"); err != nil {
		if _, err := sshclient.RunCommand(client, "sudo apt-get update && sudo apt-get install -y wireguard"); err != nil {
			return fmt.Errorf("Install WireGuard: %w", err)
		}
	}

	// Create wg0 when it does not exist yet
	if _, err := sshclient.RunCommand(client, "ip link show wg0"); err != nil {
		if err := createServerInterface(client, settings); err != nil {
			return err
		}
	}

	// Make sure wg0 starts at boot
	if _, err := sshclient.RunCommand(client, "sudo systemctl enable wg-quick@wg0"); err != nil {
		return fmt.Errorf("Enable wg0: %w", err)
	}

	return nil
}

// SetupSudoers allows the SSH user to run wg and wg-quick without a password.
func SetupSudoers() error {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return fmt.Errorf("Read settings: %w", err)
	}

	if !usernamePattern.MatchString(settings.SSHUsername) {
		return fmt.Errorf("Invalid SSH username: %q", settings.SSHUsername)
	}

	client, err := sshclient.ConnectServer()
	if err != nil {
		return err
	}
	defer client.Close()

	cmd := fmt.Sprintf(
		"echo '%s ALL=(ALL) NOPASSWD: /usr/bin/wg, /usr/bin/wg-quick' | sudo tee /etc/sudoers.d/vpn-manager && sudo chmod 440 /etc/sudoers.d/vpn-manager",
		settings.SSHUsername,
	)
	if _, err := sshclient.RunCommand(client, cmd); err != nil {
		return fmt.Errorf("Setup sudoers: %w", err)
	}

	return nil
}

// createServerInterface generates the server keys and writes the wg0 config.
func createServerInterface(client *ssh.Client, settings data.Settings) error {
	keys, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	serverAddress, err := serverAddress(settings.VPNSubnet)
	if err != nil {
		return err
	}

	conf := fmt.Sprintf(
		"[Interface]\nPrivateKey = %s\nAddress = %s\nListenPort = %s\n",
		keys.Private, serverAddress, settings.WireGuardPort,
	)

	cmd := fmt.Sprintf("sudo mkdir -p /etc/wireguard && echo '%s' | sudo tee /etc/wireguard/wg0.conf", conf)
	if _, err := sshclient.RunCommand(client, cmd); err != nil {
		return fmt.Errorf("Write wg0 config: %w", err)
	}

	if _, err := sshclient.RunCommand(client, "sudo wg-quick up wg0"); err != nil {
		return fmt.Errorf("Start wg0: %w", err)
	}

	settings.ServerPublicKey = keys.Public
	if err := data.WriteSettingsFile(settings); err != nil {
		return fmt.Errorf("Save settings: %w", err)
	}

	return nil
}

// serverAddress returns the interface address (network + 1) for a subnet.
func serverAddress(subnet string) (string, error) {
	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("Invalid VPN subnet: %w", err)
	}

	base := network.IP.To4()
	if base == nil {
		return "", fmt.Errorf("Only IPv4 subnets are supported")
	}

	ones, _ := network.Mask.Size()
	addr := net.IPv4(base[0], base[1], base[2], base[3]+1).String()

	return fmt.Sprintf("%s/%d", addr, ones), nil
}
