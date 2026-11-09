package wireguard

import (
	"errors"
	"fmt"
	"os"

	"vpn-manager/internal/data"
	sshclient "vpn-manager/internal/ssh"
)

// ErrServerConnectionFailed is returned when the server cannot be reached
// during a device delete, so the UI can offer a "delete local JSON only" option.
var ErrServerConnectionFailed = errors.New("Server connection failed")

// DeleteDevice removes a device's peer from the server and deletes its config.
func DeleteDevice(device data.Device) error {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return fmt.Errorf("Read settings: %w", err)
	}

	client, err := sshclient.ConnectServer()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerConnectionFailed, err)
	}
	defer client.Close()

	if err := RemovePeer(client, settings.ServerInterface, device.PublicKey); err != nil {
		return err
	}

	// The config may not exist yet, so a missing file is not an error.
	_ = os.Remove(data.ConfPath(device.Name))

	return nil
}
