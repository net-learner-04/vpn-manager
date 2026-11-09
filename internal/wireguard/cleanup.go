package wireguard

import (
	"os"
	"strings"

	"vpn-manager/internal/data"

	"golang.org/x/sys/windows/svc/mgr"
)

// CleanupOrphanServices removes tunnel services left behind by a previous
// abnormal exit. A service is an orphan when its config file no longer exists
// or its device is no longer in the device list.
func CleanupOrphanServices() error {
	devices, err := data.ReadDevicesFile()
	if err != nil {
		return err
	}

	knownNames := make(map[string]bool, len(devices))
	for _, d := range devices {
		knownNames[strings.ToLower(d.Name)] = true
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	names, err := m.ListServices()
	if err != nil {
		return err
	}

	for _, name := range names {
		if !strings.HasPrefix(name, "WireGuardTunnel$") {
			continue
		}

		service, err := m.OpenService(name)
		if err != nil {
			continue
		}

		config, err := service.Config()
		service.Close()
		if err != nil {
			continue
		}

		// Only consider services that point at our own conf directory, so
		// tunnels belonging to other applications are left untouched.
		if !strings.Contains(config.BinaryPathName, data.ConfDir()) {
			continue
		}

		tunnelName := strings.TrimPrefix(name, "WireGuardTunnel$")
		confPath := data.ConfPath(tunnelName)

		_, statErr := os.Stat(confPath)
		isOrphan := !knownNames[strings.ToLower(tunnelName)] || os.IsNotExist(statErr)
		if !isOrphan {
			continue
		}

		// Stop and delete the orphan service, then remove a leftover config.
		_ = DisconnectWireguard(confPath)
		if statErr == nil {
			_ = os.Remove(confPath)
		}
	}

	return nil
}
