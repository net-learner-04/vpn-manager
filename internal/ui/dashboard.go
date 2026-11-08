package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var connectionStatus *widget.Label

func BuildDashboard() fyne.CanvasObject {
	title := widget.NewLabel("Dashboard")
	title.TextStyle = fyne.TextStyle{Bold: true}

	connectionStatus = widget.NewLabel("Connection: Disconnected")
	peers := widget.NewLabel("Peers: 0")

	download := widget.NewLabel("Download: 0 B/s")
	upload := widget.NewLabel("Upload: 0 B/s")

	lasthandshake := widget.NewLabel("Last Handshake: N/A")
	totaltraffic := widget.NewLabel("Total Traffic: 0 B")

	// Connection related information gets its own card, same as the Settings page
	// Blocks are divided by the card shadow instead of separator lines
	connectionCard := widget.NewCard("Connection", "", container.NewVBox(
		connectionStatus,
		peers,
		lasthandshake,
	))

	// Transfer related information gets a separate card next to it
	transferCard := widget.NewCard("Transfer", "", container.NewVBox(
		download,
		upload,
		totaltraffic,
	))

	// Place the two cards side by side
	statusRow := container.NewGridWithColumns(2,
		connectionCard,
		transferCard,
	)

	// Create a vertical box container to hold
	// the title and the status cards
	header := container.NewVBox(
		title,
		widget.NewSeparator(),
		statusRow,
	)

	// The graph takes all the remaining space below the header
	dashboard := container.NewBorder(header, nil, nil, nil, buildGraphs())
	return dashboard
}

func buildGraphs() fyne.CanvasObject {
	// Stack layout stretches every object to the card size,
	// so a custom layout places the lines by ratio instead
	graphBackground := canvas.NewRectangle(
		color.NRGBA{R: 214, G: 211, B: 206, A: 255},
	)

	// Grid lines (kept faint so they don't look noisy)
	gridColor := color.NRGBA{R: 200, G: 197, B: 191, A: 255}
	horizontalCount := 3
	verticalCount := 0

	graphObjects := []fyne.CanvasObject{graphBackground}

	// Horizontal grid lines
	for i := 0; i < horizontalCount; i++ {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		graphObjects = append(graphObjects, line)
	}

	// Vertical grid lines
	for i := 0; i < verticalCount; i++ {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		graphObjects = append(graphObjects, line)
	}

	// Traffic line (temporary placeholder)
	trafficLine := canvas.NewLine(
		color.NRGBA{R: 71, G: 125, B: 82, A: 255},
	)
	trafficLine.StrokeWidth = 2
	graphObjects = append(graphObjects, trafficLine)

	graphCanvas := container.New(
		&graphLayout{hLines: horizontalCount, vLines: verticalCount},
		graphObjects...,
	)

	graphCard := widget.NewCard("Traffic", "", graphCanvas)

	return graphCard
}

// graphLayout places the grid lines and the traffic line relative to the graph size
type graphLayout struct {
	hLines int
	vLines int
}

func (g *graphLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(200, 150)
}

func (g *graphLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	// The first object is the background
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)

	index := 1

	// Horizontal grid lines are spread evenly over the height
	for i := 1; i <= g.hLines; i++ {
		y := size.Height * float32(i) / float32(g.hLines+1)
		line := objects[index].(*canvas.Line)
		line.Position1 = fyne.NewPos(0, y)
		line.Position2 = fyne.NewPos(size.Width, y)
		index++
	}

	// Vertical grid lines are spread evenly over the width
	for i := 1; i <= g.vLines; i++ {
		x := size.Width * float32(i) / float32(g.vLines+1)
		line := objects[index].(*canvas.Line)
		line.Position1 = fyne.NewPos(x, 0)
		line.Position2 = fyne.NewPos(x, size.Height)
		index++
	}

	// The last object is the traffic line
	trafficLine := objects[index].(*canvas.Line)
	trafficLine.Position1 = fyne.NewPos(0, size.Height*0.6)
	trafficLine.Position2 = fyne.NewPos(size.Width, size.Height*0.35)
}
