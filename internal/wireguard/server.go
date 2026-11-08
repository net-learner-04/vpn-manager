package wireguard

import (
	"errors"
	"fmt"
	"net"
	"strings"

	sshclient "vpn-manager/internal/ssh"

	"golang.org/x/crypto/ssh"
)

// Get the server's WireGuard public key over SSH
func GetServerPublicKey(client *ssh.Client, iface string) (string, error) {
	if err := validateInterface(iface); err != nil {
		return "", err
	}

	output, err := sshclient.RunCommand(client, fmt.Sprintf("sudo wg show %s public-key", iface))
	if err != nil {
		return "", fmt.Errorf("Get server public key: %w - %s", err, output)
	}

	key := strings.TrimSpace(output)
	if err := validatePublicKey(key); err != nil {
		return "", fmt.Errorf("Get server public key: %w", err)
	}

	return key, nil
}

// Get the IPs already assigned to peers on the server (without masks)
func GetUsedIPs(client *ssh.Client, iface string) (map[string]bool, error) {
	if err := validateInterface(iface); err != nil {
		return nil, err
	}

	output, err := sshclient.RunCommand(client, fmt.Sprintf("sudo wg show %s allowed-ips", iface))
	if err != nil {
		return nil, fmt.Errorf("Get used IPs: %w - %s", err, output)
	}

	used := make(map[string]bool)

	// Each line looks like: <public key>\t10.0.0.2/32 10.0.0.3/32
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		for _, cidr := range fields[1:] {
			ip, _, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			used[ip.String()] = true
		}
	}

	return used, nil
}

// Find the next free IP in the VPN subnet.
func AllocateIP(client *ssh.Client, iface, subnet string, extraUsed []string) (string, error) {
	used, err := GetUsedIPs(client, iface)
	if err != nil {
		return "", err
	}

	for _, ip := range extraUsed {
		used[ip] = true
	}

	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("Invalid VPN subnet: %w", err)
	}

	// Start from the network address and walk the subnet
	base := network.IP.To4()
	if base == nil {
		return "", errors.New("Only IPv4 subnets are supported")
	}

	ones, bits := network.Mask.Size()
	size := 1 << uint(bits-ones)

	for i := 2; i < size-1; i++ {
		candidate := net.IPv4(base[0], base[1], base[2], base[3]+byte(i))
		if !used[candidate.String()] {
			return candidate.String(), nil
		}
	}

	return "", errors.New("No free IP addresses left in the VPN subnet")
}
