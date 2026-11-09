package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync"
	"time"

	"vpn-manager/internal/data"
	"vpn-manager/internal/wireguard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	connectionStatus *widget.Label
	peersLabel       *widget.Label
	downloadLabel    *widget.Label
	uploadLabel      *widget.Label
	handshakeLabel   *widget.Label
	totalLabel       *widget.Label

	trafficGraphWidget *trafficGraph
	statsErrorLogged   bool
	lastRx             uint64
	lastTx             uint64
	lastStatsTime      time.Time
	trafficSamples     []float64
)

func BuildDashboard() fyne.CanvasObject {
	title := widget.NewLabel("Dashboard")
	title.TextStyle = fyne.TextStyle{Bold: true}

	connectionStatus = widget.NewLabel("Connection: Disconnected")
	peersLabel = widget.NewLabel("Peers: 0")

	downloadLabel = widget.NewLabel("Download: 0 B/s")
	uploadLabel = widget.NewLabel("Upload: 0 B/s")

	handshakeLabel = widget.NewLabel("Last Handshake: N/A")
	totalLabel = widget.NewLabel("Total Traffic: 0 B")

	// Connection related information gets its own card, same as the Settings page
	// Blocks are divided by the card shadow instead of separator lines
	connectionCard := widget.NewCard("Connection", "", container.NewVBox(
		connectionStatus,
		peersLabel,
		handshakeLabel,
	))

	// Transfer related information gets a separate card next to it
	transferCard := widget.NewCard("Transfer", "", container.NewVBox(
		downloadLabel,
		uploadLabel,
		totalLabel,
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
	trafficGraphWidget = newTrafficGraph()
	dashboard := container.NewBorder(header, nil, nil, nil, widget.NewCard("Traffic", "", trafficGraphWidget))
	return dashboard
}

// RefreshDashboard updates the dashboard labels with real data.
func RefreshDashboard() {
	// Peer count comes from the local device list
	if devices, err := data.ReadDevicesFile(); err == nil && peersLabel != nil {
		peersLabel.SetText(fmt.Sprintf("Peers: %d", len(devices)))
	}

	// Server statistics are queried in the background so the UI stays responsive
	go func() {
		stats, err := wireguard.GetServerStats()
		if err != nil {
			// Log the failure once so the Logs tab shows why stats stay empty
			if !statsErrorLogged {
				statsErrorLogged = true
				writeLog("WARN", "Failed to read server stats: "+err.Error())
			}
			return
		}
		statsErrorLogged = false

		now := time.Now()
		var rxRate, txRate float64
		if !lastStatsTime.IsZero() {
			elapsed := now.Sub(lastStatsTime).Seconds()
			if elapsed > 0 {
				if stats.TotalRx >= lastRx {
					rxRate = float64(stats.TotalRx-lastRx) / elapsed
				}
				if stats.TotalTx >= lastTx {
					txRate = float64(stats.TotalTx-lastTx) / elapsed
				}
			}
		}
		lastRx, lastTx, lastStatsTime = stats.TotalRx, stats.TotalTx, now

		// Download = bytes the client receives (server tx), upload = bytes sent
		downloadRate, uploadRate := txRate, rxRate

		trafficSamples = append(trafficSamples, rxRate+txRate)
		if len(trafficSamples) > 60 {
			trafficSamples = trafficSamples[1:]
		}
		samples := append([]float64(nil), trafficSamples...)

		fyne.Do(func() {
			if downloadLabel == nil {
				return
			}

			downloadLabel.SetText("Download: " + formatRate(downloadRate))
			uploadLabel.SetText("Upload: " + formatRate(uploadRate))
			totalLabel.SetText("Total Traffic: " + formatBytes(stats.TotalRx+stats.TotalTx))

			if stats.LastHandshake.IsZero() {
				handshakeLabel.SetText("Last Handshake: N/A")
			} else {
				handshakeLabel.SetText("Last Handshake: " + stats.LastHandshake.Format("15:04:05"))
			}

			if trafficGraphWidget != nil {
				trafficGraphWidget.setValues(samples)
			}
		})
	}()
}

// StartDashboardRefresh refreshes the dashboard labels on a timer.
func StartDashboardRefresh() {
	go func() {
		RefreshDashboard()
		for {
			time.Sleep(refreshInterval())
			RefreshDashboard()
		}
	}()
}

// refreshInterval returns the configured dashboard refresh interval.
func refreshInterval() time.Duration {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return 5 * time.Second
	}

	seconds, err := strconv.Atoi(settings.RefreshInterval)
	if err != nil || seconds <= 0 {
		return 5 * time.Second
	}

	return time.Duration(seconds) * time.Second
}

// formatBytes formats a byte count in a human readable way.
func formatBytes(n uint64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.2f GiB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.2f MiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.2f KiB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// formatRate formats a transfer rate in a human readable way.
func formatRate(rate float64) string {
	switch {
	case rate >= 1<<30:
		return fmt.Sprintf("%.2f GiB/s", rate/(1<<30))
	case rate >= 1<<20:
		return fmt.Sprintf("%.2f MiB/s", rate/(1<<20))
	case rate >= 1<<10:
		return fmt.Sprintf("%.2f KiB/s", rate/(1<<10))
	default:
		return fmt.Sprintf("%.0f B/s", rate)
	}
}

// trafficGraph draws a moving line of recent throughput samples.
type trafficGraph struct {
	widget.BaseWidget

	mu     sync.Mutex
	values []float64
	max    float64
}

func newTrafficGraph() *trafficGraph {
	g := &trafficGraph{}
	g.ExtendBaseWidget(g)
	return g
}

func (g *trafficGraph) setValues(values []float64) {
	g.mu.Lock()
	g.values = append([]float64(nil), values...)
	g.max = 0
	for _, v := range g.values {
		if v > g.max {
			g.max = v
		}
	}
	if g.max <= 0 {
		g.max = 1
	}
	g.mu.Unlock()
	g.Refresh()
}

func (g *trafficGraph) CreateRenderer() fyne.WidgetRenderer {
	return &trafficGraphRenderer{
		g:      g,
		raster: canvas.NewRaster(g.draw),
	}
}

func (g *trafficGraph) draw(w, h int) image.Image {
	g.mu.Lock()
	values := append([]float64(nil), g.values...)
	max := g.max
	g.mu.Unlock()

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	background := color.RGBA{R: 214, G: 211, B: 206, A: 255}
	grid := color.RGBA{R: 200, G: 197, B: 191, A: 255}
	line := color.RGBA{R: 71, G: 125, B: 82, A: 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, background)
		}
	}

	for i := 1; i <= 3; i++ {
		y := h * i / 4
		for x := 0; x < w; x++ {
			img.Set(x, y, grid)
		}
	}

	if w <= 1 || len(values) < 2 {
		return img
	}

	for x := 0; x < w; x++ {
		idx := x * (len(values) - 1) / (w - 1)
		v := values[idx]
		y := h - int(float64(h)*v/max)
		if y < 0 {
			y = 0
		}
		if y >= h {
			y = h - 1
		}
		img.Set(x, y, line)
		if y > 0 {
			img.Set(x, y-1, line)
		}
	}

	return img
}

type trafficGraphRenderer struct {
	g      *trafficGraph
	raster *canvas.Raster
}

func (r *trafficGraphRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *trafficGraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 150)
}

func (r *trafficGraphRenderer) Refresh() {
	r.raster.Refresh()
}

func (r *trafficGraphRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *trafficGraphRenderer) Destroy() {}
