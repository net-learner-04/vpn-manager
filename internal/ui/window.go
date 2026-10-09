package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildWindow(w fyne.Window) {
	// Set the window size to 700x200 pixels.
	w.Resize(fyne.NewSize(600, 400))

	title := widget.NewLabel("VPN Manager")
	title.TextStyle = fyne.TextStyle{Bold: true}

	status := widget.NewLabel("Status: Disconnected")
	peers := widget.NewLabel("Peers: 0")
	download := widget.NewLabel("Download: 0 B/s")
	upload := widget.NewLabel("Upload: 0 B/s")

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			status.SetText("Status: clicked refresh button")
		},
	)

	testButton := widget.NewButton(
		"Server Connection Test",
		func() {
			status.SetText("Status: clicked test button")
		},
	)

	addPeerButton := widget.NewButton(
		"Add a New Device",
		func() {
			status.SetText("Status: clicked add peer button")
		},
	)

	// Create a vertical box container to hold
	// the title, status, peers, download, upload, and buttons.
	dashboard := container.NewVBox(
		title,
		// Add a separator line between the title and the status.
		widget.NewSeparator(),

		container.NewGridWithColumns(2,
			status,
			peers,
		),

		container.NewGridWithColumns(2,
			download,
			upload,
		),

		widget.NewSeparator(),

		container.NewVBox(
			refreshButton,
			testButton,
			addPeerButton,
		),
	)

	menuItems := []string{
		"Dashboard",
		"Devices",
		"Traffic",
		"Logs",
		"Settings",
	}

	menu := widget.NewList(
		// Return the number of items in the list.
		func() int {
			return len(menuItems)
		},
		// Create a new label for each item in the list.
		func() fyne.CanvasObject {
			return widget.NewLabel("Menu")
		},
		// Update the label for each item in the list.
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(menuItems[id])
		},
	)

	split := container.NewHSplit(
		menu,
		container.NewPadded(dashboard),
	)

	// Set the initial position of the split to 20%
	// for the menu and 80% for the dashboard.
	split.SetOffset(0.2)

	w.SetContent(split)
}
