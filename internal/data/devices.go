package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// Specify the file path where you want to save it later
const devicesfile = "internal/data/devices.json"

type Device struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	PublicKey string `json:"public_key"`
}

func CreateDevicesFile() error {
	_, err := os.Stat(devicesfile)
	if err == nil {
		// The file already exists
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		// Errors Other Than “File Not Found”
		return err
	}

	// If the file doesn't exist, create it
	if err := os.WriteFile(devicesfile, []byte("[]"), 0644); err != nil {
		return err
	}

	return nil
}

func ReadDevicesFile() ([]Device, error) {
	filedata, err := os.ReadFile(devicesfile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := CreateDevicesFile(); err != nil {
				return nil, err
			}

			filedata, err = os.ReadFile(devicesfile)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var data []Device
	if err := json.Unmarshal(filedata, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func WriteDevicesFile(devices []Device) error {
	jsondata, err := json.MarshalIndent(devices, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(devicesfile, jsondata, 0644)
	if err != nil {
		return err
	}

	return nil
}
