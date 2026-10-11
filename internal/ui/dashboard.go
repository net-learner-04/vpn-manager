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
	statusCard := widget.NewCard("", "", statusContainer)

	refreshButton := widget.NewButton(
		"Refresh Status",
		func() {
			connectionStatus.SetText("Connection: clicked refresh button")
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

		statusCard,

		widget.NewSeparator(),

		container.NewVBox(
			refreshButton,
			addPeerButton,
		),

		widget.NewSeparator(),

		buildGraphs(),
	)
	return dashboard
}

func buildGraphs() fyne.CanvasObject {
	graphWidth := float32(550)
	graphHeight := float32(150)

	graphCanvas := container.NewWithoutLayout()
	graphCanvas.Resize(fyne.NewSize(graphWidth, graphHeight))

	graphBackgroun := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 40, A: 255})
	graphBackgroun.Resize(fyne.NewSize(graphWidth, graphHeight))
	graphCanvas.Add(graphBackgroun)

	line := canvas.NewLine(color.RGBA{R: 0, G: 180, B: 255, A: 255})
	line.StrokeWidth = 2

	line.Resize(fyne.NewSize(graphWidth, graphHeight))
	graphCanvas.Add(line)

	graphCard := widget.NewCard("Network Traffic Graph", "", graphCanvas)

	return graphCard
}
