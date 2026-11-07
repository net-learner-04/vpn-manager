package wireguard

import (
	"errors"
	"fmt"
	"os/exec"
)

func ConnectWireguard() error {
	if installed, err := IsWireGuardInstalled(); err != nil {
		return err
	} else if !installed {
		// WireGuard is not installed, cannot connect
		err := errors.New("WireGuard is not installed")
		return err
	}

	// Start the WireGuard service using the Windows Service Control Manager (sc.exe)
	if active, err := IsWireGuardRunning(); err != nil {
		return err
	} else if active {
		// WireGuard is already running, no need to start it again
		return nil
	}

	// Start the WireGuard service using the Windows Service Control Manager (sc.exe)
	cmd := exec.Command("sc.exe", "start", "WireGuardTunnel$vpn-manager")
	output, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to start WireGuard: %w - %s", err, string(output))
		return err
	}

	return nil
}

func DisconnectWireguard() error {
	// Stop the WireGuard service using the Windows Service Control Manager (sc.exe)
	if active, err := IsWireGuardRunning(); err != nil {
		return err
	} else if !active {
		// WireGuard is not running, no need to stop it
		return nil
	}

	// Stop the WireGuard service using the Windows Service Control Manager (sc.exe)
	cmd := exec.Command("sc.exe", "stop", "WireGuardTunnel$vpn-manager")
	output, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to stop WireGuard: %w - %s", err, string(output))
		return err
	}

	return nil
}
