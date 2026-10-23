package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

	// Group status information into a card for a cleaner dashboard look
	statusContainer := container.NewVBox(
		serverStatus,
		widget.NewSeparator(),
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
	)
	statusCard := widget.NewCard("Status", "", statusContainer)

	connectButton := widget.NewButton(
		"Connect",
		func() {
			connectionStatus.SetText("Connection: Connected")
		},
	)
	connectButton.Importance = widget.HighImportance

	unconnectButton := widget.NewButton(
		"Unconnected",
		func() {
			connectionStatus.SetText("Connection: Unconnected")
		},
	)
	unconnectButton.Importance = widget.DangerImportance

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			connectionStatus.SetText("Connection: clicked refresh button")
		},
	)

	// Create a vertical box container to hold
	// the title, status, peers, download, upload, and buttons
	header := container.NewVBox(
		title,

		widget.NewSeparator(),

		statusCard,
	)

	// Buttons sit at the bottom, same as the Settings page
	footer := container.NewVBox(
		widget.NewSeparator(),

		container.NewHBox(
			connectButton,
			unconnectButton,
			refreshButton,
		),
	)

	// The graph takes all the remaining space between header and footer
	dashboard := container.NewBorder(header, footer, nil, nil, buildGraphs())
	return dashboard
}

func buildGraphs() fyne.CanvasObject {
	graphHeight := float32(150)

	// Stack layout makes the background and line follow the card size
	graphBackgroun := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 40, A: 255})
	graphBackgroun.SetMinSize(fyne.NewSize(0, graphHeight))

	line := canvas.NewLine(color.RGBA{R: 0, G: 220, B: 0, A: 255})
	line.StrokeWidth = 2

	graphCanvas := container.NewStack(graphBackgroun, line)

	graphCard := widget.NewCard("Traffic", "", graphCanvas)

	return graphCard
}
