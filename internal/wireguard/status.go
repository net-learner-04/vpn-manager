package wireguard

import (
	"errors"
	"os/exec"
	"strings"
)

func IsWireGuardRunning() (bool, error) {
	// Check the status of the WireGuard service using the Windows Service Control Manager (sc.exe)
	cmd := exec.Command("sc.exe", "query", "WireGuardTunnel$vpn-manager")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}

	// Check if the output contains the "RUNNING" status
	if strings.Contains(string(output), "RUNNING") {
		return true, nil
	}

	return false, nil
}

func IsWireGuardInstalled() (bool, error) {
	// Check the status of the WireGuard service using the Windows Service Control Manager (sc.exe)
	cmd := exec.Command("sc.exe", "query", "WireGuardTunnel$vpn-manager")
	_, err := cmd.Output()

	if err != nil {
		// Check if the service does not exist (error code 1060)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1060 {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
