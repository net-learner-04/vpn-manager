package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildDashboard() fyne.CanvasObject {
	title := widget.NewLabel("Dashboard")
	title.TextStyle = fyne.TextStyle{Bold: true}

	serverStatus := widget.NewLabel("Server Status: Unknown")
	connectionStatus := widget.NewLabel("Connection: Disconnected")
	peers := widget.NewLabel("Peers: 0")
	download := widget.NewLabel("Download: 0 B/s")
	upload := widget.NewLabel("Upload: 0 B/s")
	lasthandshake := widget.NewLabel("Last Handshake: N/A")
	totaltraffic := widget.NewLabel("Total Traffic: 0 B")

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			connectionStatus.SetText("Connection: clicked refresh button")
		},
	)

	testButton := widget.NewButton(
		"Server Connection Test",
		func() {
			connectionStatus.SetText("Connection: clicked test button")
		},
	)

	addPeerButton := widget.NewButton(
		"Add a New Device",
		func() {
			connectionStatus.SetText("Connection: clicked add peer button")
		},
	)

	// Create a vertical box container to hold
	// the title, status, peers, download, upload, and buttons.
	dashboard := container.NewVBox(
		title,
		// Add a separator line between the title and the status.
		widget.NewSeparator(),

		serverStatus,

		container.NewGridWithColumns(2,
			connectionStatus,
			peers,
		),

		container.NewGridWithColumns(2,
			download,
			upload,
		),

		container.NewGridWithColumns(2,
			lasthandshake,
			totaltraffic,
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
