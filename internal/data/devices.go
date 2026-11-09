package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Device struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	PublicKey string `json:"public_key"`
}

func CreateDevicesFile() error {
	path := DevicesPath()

	_, err := os.Stat(path)
	if err == nil {
		// The file already exists
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		// Errors Other Than “File Not Found”
		return err
	}

	// Migrate the legacy devices file if it still exists
	if migrateLegacyFile(legacyDevicesFile, path) {
		return nil
	}

	// Create the parent directory if it does not exist yet
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// If the file doesn't exist, create it
	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		return err
	}

	return nil
}

func ReadDevicesFile() ([]Device, error) {
	filedata, err := os.ReadFile(DevicesPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := CreateDevicesFile(); err != nil {
				return nil, err
			}

			filedata, err = os.ReadFile(DevicesPath())
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

	err = os.WriteFile(DevicesPath(), jsondata, 0644)
	if err != nil {
		return err
	}

	return nil
}
