package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildDashboard() fyne.CanvasObject {
	title := widget.NewLabel("Dashboard")
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
	return dashboard
}
