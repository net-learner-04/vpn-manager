package ui

import (
	"vpn-manager/internal/data"
	"vpn-manager/internal/wireguard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildDevicesItem(device data.Device, status string, w fyne.Window,
	Delete func(data.Device)) fyne.CanvasObject {
	name := widget.NewLabel(device.Name)
	name.TextStyle = fyne.TextStyle{Bold: true}

	ip := widget.NewLabel(device.IP)
	statusLabel := widget.NewLabel(status)

	deleteButton := widget.NewButtonWithIcon(
		"",
		theme.DeleteIcon(),
		func() {
			dialog.NewConfirm(
				"Delete Device",
				"Do you want to delete this device?",
				func(b bool) {
					if b {
						Delete(device)
					}
				},
				w,
			).Show()
		},
	)
	deleteButton.Importance = widget.LowImportance

	right := container.NewCenter(
		container.NewHBox(statusLabel, deleteButton),
	)

	return container.NewBorder(
		// top, bottom, left, right, center
		nil,
		nil,
		nil,
		right,
		container.NewVBox(name, ip),
	)
}

func BuildDevices(w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Devices")
	title.TextStyle = fyne.TextStyle{Bold: true}

	deviceList, err := data.ReadDevicesFile()
	if err != nil {
		writeLog("ERROR", "failed to read devices file: "+err.Error())
		dialog.ShowError(err, w)

		return container.NewVBox(
			widget.NewLabel("The device list cannot be loaded"),
		)
	}

	writeLog("INFO", "Device list loaded successfully")

	// Creating a Container List
	listContainer := container.NewVBox()

	var Delete func(data.Device)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search devices...")

	refreshDeviceList := func() {
		listContainer.RemoveAll()

		matched := FilterDevices(deviceList, searchEntry.Text)

		for _, device := range matched {
			item := BuildDevicesItem(device, "test", w, Delete)
			listContainer.Add(item)
			listContainer.Add(widget.NewSeparator())
		}

		listContainer.Refresh()

		writeLog("INFO", "Device list refreshed")
	}

	searchEntry.OnChanged = func(s string) {
		refreshDeviceList()
	}

	Delete = func(device data.Device) {
		// Save only devices that are not scheduled for deletion to the new list
		newList := []data.Device{}

		for _, target := range deviceList {
			if !(target.Name == device.Name &&
				target.IP == device.IP &&
				target.PublicKey == device.PublicKey) {
				newList = append(newList, target)
			}
		}

		// Save JSON
		if err := data.WriteDevicesFile(newList); err != nil {
			writeLog("ERROR", "failed to write devices file: "+err.Error())
			dialog.ShowError(err, w)
			return
		}

		writeLog("INFO", "Device deleted successfully: "+device.Name)

		// Memory Refresh
		deviceList = newList

		// Refresh the screen list
		refreshDeviceList()
	}

	addButton := widget.NewButton(
		"Add New Device",
		func() {
			name := widget.NewEntry()
			name.SetPlaceHolder("Enter The Device Name...")

			dialog.NewCustomConfirm(
				"Add New Device",
				"Add",
				"Cancel",
				container.NewVBox(name),
				func(b bool) {
					if !b {
						return
					}

					deviceName := strings.TrimSpace(name.Text)

					// Check required fields
					if deviceName == "" {
						writeLog("WARN", "Device name is missing")
						dialog.ShowError(fmt.Errorf("device name is required"), w)
						return
					}

					// Run in a goroutine so the UI thread is not blocked
					go func() {
						// Create the device and save it to the devices file
						// confDir is set to a temporary directory because the config file is not needed for the UI
						newDevice, err := wireguard.CreateDevice(deviceName, filepath.Join(os.TempDir(), "vpn-manager"))
						if err != nil {
							writeLog("ERROR", "failed to create device: "+err.Error())
							dialog.ShowError(err, w)
							return
						}

						updatedDevices := append(
							append([]data.Device(nil), deviceList...),
							newDevice,
						)

						if err := data.WriteDevicesFile(updatedDevices); err != nil {
							writeLog("ERROR", "failed to write devices file: "+err.Error())
							dialog.ShowError(err, w)
							return
						}

						writeLog("INFO", "Device added successfully: "+deviceName)

						deviceList = updatedDevices
						refreshDeviceList()
					}()
				},
				w,
			).Show()
		},
	)

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			refreshDeviceList()
		},
	)

	footer := container.NewGridWithColumns(
		2,
		addButton,
		refreshButton,
	)

	devices := container.NewBorder(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			searchEntry,
		),
		container.NewVBox(
			widget.NewSeparator(),
			footer,
		),
		nil,
		container.NewVScroll(listContainer),
	)

	return devices
}

func FilterDevices(devices []data.Device, keyword string) []data.Device {
	result := []data.Device{}

	keyword = strings.ToLower(keyword)

	for _, device := range devices {
		name := strings.ToLower(device.Name)
		ip := strings.ToLower(device.IP)

		if strings.Contains(name, keyword) ||
			strings.Contains(ip, keyword) {
			result = append(result, device)
		}
	}

	return result
}
