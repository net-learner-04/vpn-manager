package wireguard

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"

	sshclient "vpn-manager/internal/ssh"

	"golang.org/x/crypto/ssh"
)

// Check that the string is a valid WireGuard key (base64, 32 bytes)
func validatePublicKey(key string) error {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(b) != 32 {
		return errors.New("Invalid WireGuard public key")
	}
	return nil
}

// Check that the interface name only has safe characters
func validateInterface(name string) error {
	if name == "" || len(name) > 15 {
		return errors.New("Invalid interface name")
	}

	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-'
		if !ok {
			return errors.New("Invalid interface name")
		}
	}

	return nil
}

// Register a client on the server.
func AddPeer(client *ssh.Client, iface, pubKey, ip string) error {
	if err := validateInterface(iface); err != nil {
		return err
	}

	if err := validatePublicKey(pubKey); err != nil {
		return err
	}

	if net.ParseIP(ip) == nil {
		return fmt.Errorf("Invalid IP address: %q", ip)
	}

	// "wg set" applies immediately, "wg-quick save" writes it to the config file
	cmd := fmt.Sprintf(
		"sudo wg set %s peer %s allowed-ips %s/32 && sudo wg-quick save %s",
		iface, pubKey, ip, iface,
	)

	if output, err := sshclient.RunCommand(client, cmd); err != nil {
		return fmt.Errorf("Add peer: %w - %s", err, output)
	}

	return nil
}

// Remove a client from the server
func RemovePeer(client *ssh.Client, iface, pubKey string) error {
	if err := validateInterface(iface); err != nil {
		return err
	}

	if err := validatePublicKey(pubKey); err != nil {
		return err
	}

	cmd := fmt.Sprintf(
		"sudo wg set %s peer %s remove && sudo wg-quick save %s",
		iface, pubKey, iface,
	)

	if output, err := sshclient.RunCommand(client, cmd); err != nil {
		return fmt.Errorf("Remove peer: %w - %s", err, output)
	}

	return nil
}
