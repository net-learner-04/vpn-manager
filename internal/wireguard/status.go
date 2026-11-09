package wireguard

import (
	"errors"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func IsWireGuardRunning(confPath string) (bool, error) {
	serviceName, err := serviceNameForPath(confPath)
	if err != nil {
		return false, err
	}

	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return false, nil
		}
		return false, err
	}
	defer service.Close()

	status, err := service.Query()
	if err != nil {
		return false, err
	}

	return status.State == svc.Running, nil
}

func IsWireGuardInstalled(confPath string) (bool, error) {
	serviceName, err := serviceNameForPath(confPath)
	if err != nil {
		return false, err
	}

	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return false, nil
		}
		return false, err
	}
	service.Close()

	return true, nil
}
