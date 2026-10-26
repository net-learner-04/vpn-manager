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
	graphWidth := float32(570)

	// Stack layout makes the background and line follow the card size
	graphBackground := canvas.NewRectangle(
		color.NRGBA{R: 214, G: 211, B: 206, A: 255},
	)
	graphBackground.SetMinSize(fyne.NewSize(graphWidth, graphHeight))

	// Grid lines
	gridColor := color.NRGBA{R: 184, G: 181, B: 174, A: 255}
	gridLines := []fyne.CanvasObject{}

	// Horizontal grid lines
	for _, y := range []float32{30, 60, 90, 120} {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		line.Position1 = fyne.NewPos(0, y)
		line.Position2 = fyne.NewPos(graphWidth, y)
		gridLines = append(gridLines, line)
	}

	// Vertical grid lines
	for _, x := range []float32{95, 190, 285, 380, 475} {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		line.Position1 = fyne.NewPos(x, 0)
		line.Position2 = fyne.NewPos(x, graphHeight)
		gridLines = append(gridLines, line)
	}

	// Traffic line (temporary placeholder)
	trafficLine := canvas.NewLine(
		color.NRGBA{R: 71, G: 125, B: 82, A: 255},
	)
	trafficLine.StrokeWidth = 2
	trafficLine.Position1 = fyne.NewPos(0, 90)
	trafficLine.Position2 = fyne.NewPos(graphWidth, 55)

	// Combine graph objects
	graphObjects := []fyne.CanvasObject{graphBackground}
	graphObjects = append(graphObjects, gridLines...)
	graphObjects = append(graphObjects, trafficLine)

	graphCanvas := container.NewStack(graphObjects...)

	graphCard := widget.NewCard("Traffic", "", graphCanvas)

	return graphCard
}
