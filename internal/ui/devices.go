package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"strings"
)

type Device struct {
	Name   string
	IP     string
	Status string
}

func BuildDevicesItem(device Device, w fyne.Window) fyne.CanvasObject {
	name := widget.NewLabel(device.Name)
	name.TextStyle = fyne.TextStyle{Bold: true}

	ip := widget.NewLabel(device.IP)
	status := widget.NewLabel(device.Status)

	deleteButton := widget.NewButtonWithIcon(
		"",
		theme.DeleteIcon(),
		func() {
			dialog.NewConfirm(
				"Delete Device",
				"Do you want to delete this device?",
				func(b bool) {
					if b {
						// to do : delete logic
					}
				},
				w,
			).Show()
		},
	)
	deleteButton.Importance = widget.LowImportance

	right := container.NewCenter(
		container.NewHBox(status, deleteButton),
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

	// for the test
	deviceList := []Device{
		{
			Name:   "My Windows PC",
			IP:     "10.0.0.2/32",
			Status: "Connected",
		},
		{
			Name:   "Laptop",
			IP:     "10.0.0.3/32",
			Status: "Connected",
		},
		{
			Name:   "Android Phone",
			IP:     "10.0.0.4/32",
			Status: "Offline",
		},
	}

	listContainer := container.NewVBox()

	for _, device := range deviceList {
		item := BuildDevicesItem(device, w)
		listContainer.Add(item)
		listContainer.Add(widget.NewSeparator())
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search devices...")
	searchEntry.OnChanged = func(s string) {
		matched := FilterDevices(deviceList, s)

		listContainer.RemoveAll()

		for _, device := range matched {
			item := BuildDevicesItem(device, w)
			listContainer.Add(item)
			listContainer.Add(widget.NewSeparator())
		}

		listContainer.Refresh()
	}

	addButton := widget.NewButton(
		"Add New Device",
		func() {
			// to do : add devices
		},
	)

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			// to do : refresh status
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

func FilterDevices(devices []Device, keyword string) []Device {
	result := []Device{}

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
