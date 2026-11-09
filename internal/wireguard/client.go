package wireguard

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"golang.zx2c4.com/wireguard/windows/conf"
)

// serviceNameForPath returns the Windows service name for a config path.
func serviceNameForPath(confPath string) (string, error) {
	name, err := conf.NameFromPath(confPath)
	if err != nil {
		return "", err
	}
	return conf.ServiceNameOfTunnel(name)
}

func ConnectWireguard(confPath string) error {
	serviceName, err := serviceNameForPath(confPath)
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return err
		}

		// Create the service if it does not exist yet
		exe, err := os.Executable()
		if err != nil {
			return err
		}

		service, err = m.CreateService(serviceName, exe, mgr.Config{
			StartType:   mgr.StartManual,
			DisplayName: serviceName,
			SidType:     windows.SERVICE_SID_TYPE_UNRESTRICTED,
		}, "/tunnelservice", confPath)
		if err != nil {
			return err
		}
	}
	defer service.Close()

	if err := ensureServiceSID(service); err != nil {
		return err
	}

	status, err := service.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Running {
		return nil
	}

	// Only send the start control when the service is not already starting.
	if status.State != svc.StartPending {
		if err := service.Start(); err != nil {
			return err
		}
	}

	return waitForRunning(service)
}

// waitForRunning waits for the tunnel service to reach the running state so a
// successful result means the tunnel actually started. The service exits on its
// own when the tunnel cannot be brought up, so a stopped state is an error.
func waitForRunning(service *mgr.Service) error {
	for i := 0; i < 40; i++ {
		time.Sleep(250 * time.Millisecond)

		status, err := service.Query()
		if err != nil {
			return err
		}

		switch status.State {
		case svc.Running:
			return nil
		case svc.Stopped:
			return errors.New("WireGuard tunnel service stopped during startup")
		}
	}

	return errors.New("WireGuard tunnel service did not start in time")
}

// ensureServiceSID gives the tunnel service an unrestricted service SID.
// WireGuard's firewall rules look for this SID in the process token and fail
// with ERROR_NO_SUCH_GROUP when it is missing.
func ensureServiceSID(service *mgr.Service) error {
	config, err := service.Config()
	if err != nil {
		return err
	}
	if config.SidType == windows.SERVICE_SID_TYPE_UNRESTRICTED {
		return nil
	}
	config.SidType = windows.SERVICE_SID_TYPE_UNRESTRICTED
	return service.UpdateConfig(config)
}

func DisconnectWireguard(confPath string) error {
	serviceName, err := serviceNameForPath(confPath)
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		// The service does not exist, so there is nothing to disconnect
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return nil
		}
		return err
	}
	defer service.Close()

	status, err := service.Query()
	if err != nil {
		return err
	}
	if status.State != svc.Stopped {
		if _, err := service.Control(svc.Stop); err != nil {
			return err
		}
	}

	return service.Delete()
}
