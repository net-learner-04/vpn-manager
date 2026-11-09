package wireguard

import (
	"errors"
	"os"

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
		}, "/tunnelservice", confPath)
		if err != nil {
			return err
		}
	}
	defer service.Close()

	status, err := service.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Running {
		return nil
	}

	return service.Start()
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
