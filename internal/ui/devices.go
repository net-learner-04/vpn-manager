package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"strings"
)

type Device struct {
	Name   string
	IP     string
	Status string
}

func BuildDeviceItem(device Device) fyne.CanvasObject {
	name := widget.NewLabel(device.Name)
	name.TextStyle = fyne.TextStyle{Bold: true}

	ip := widget.NewLabel(device.IP)
	status := widget.NewLabel(device.Status)

	details := container.NewVBox(
		name,
		ip,
	)

	return container.NewBorder(
		nil, nil, nil, status,
		details,
	)
}

func BuildDevices() fyne.CanvasObject {
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
		item := BuildDeviceItem(device)
		listContainer.Add(item)
		listContainer.Add(widget.NewSeparator())
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search devices...")
	searchEntry.OnChanged = func(s string) {
		matched := FilterDevices(deviceList, s)

		listContainer.RemoveAll()

		for _, device := range matched {
			item := BuildDeviceItem(device)
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
	addButton.Importance = widget.HighImportance

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			// to do : refresh status
		},
	)
	refreshButton.Importance = widget.SuccessImportance

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
