package ui

import (
	"vpn-manager/internal/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fmt"
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

	deviceList, err := data.ReadJsonFile()
	if err != nil {
		dialog.ShowError(err, w)

		return container.NewVBox(
			widget.NewLabel("The device list cannot be loaded"),
		)
	}

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
		if err := data.WriteJsonFile(newList); err != nil {
			dialog.ShowError(err, w)
			return
		}

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

			ip := widget.NewEntry()
			ip.SetPlaceHolder("Enter The IP Address...")

			publickey := widget.NewEntry()
			publickey.SetPlaceHolder("Enter The PublicKey...")

			form := container.NewVBox(
				name,
				ip,
				publickey,
			)

			dialog.NewCustomConfirm(
				"Add New Device",
				"Add",
				"Cancel",
				form,
				func(b bool) {
					if !b {
						return
					}

					deviceName := strings.TrimSpace(name.Text)
					deviceIp := strings.TrimSpace(ip.Text)
					devicePublickey := strings.TrimSpace(publickey.Text)

					// Check required fields
					if deviceName == "" || deviceIp == "" {
						dialog.ShowError(
							fmt.Errorf("device name and IP address are required"),
							w,
						)
						return
					}

					// to do : Enhanced validation logic,
					// such as IP address format validation
					// and public key format validation

					newDevice := data.Device{
						Name:      deviceName,
						IP:        deviceIp,
						PublicKey: devicePublickey,
					}

					// Create a new list without changing the original
					updatedDevices := append(
						append([]data.Device(nil), deviceList...),
						newDevice,
					)

					// Save JSON first
					if err := data.WriteJsonFile(updatedDevices); err != nil {
						dialog.ShowError(err, w)
						return
					}

					// Update in-memory list after successful save
					deviceList = updatedDevices

					// Refresh the device list
					listContainer.Objects = nil

					filteredDevices := FilterDevices(deviceList, searchEntry.Text)

					for _, device := range filteredDevices {
						item := BuildDevicesItem(device, "test", w, Delete)
						listContainer.Add(item)
						listContainer.Add(widget.NewSeparator())
					}

					listContainer.Refresh()
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

	devices := container.NewBorder(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			searchEntry,
		),
		container.NewVBox(
			widget.NewSeparator(),
			addButton,
			refreshButton,
		),
		nil,
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
}
