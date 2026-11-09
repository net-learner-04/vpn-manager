package ui

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"vpn-manager/internal/data"
	"vpn-manager/internal/wireguard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	connectionStatus *widget.Label
	tunnelConnected  atomic.Bool
	peersLabel       *widget.Label
	downloadLabel    *widget.Label
	uploadLabel      *widget.Label
	handshakeLabel   *widget.Label
	totalLabel       *widget.Label

	statsErrorLogged bool
	lastRx           uint64
	lastTx           uint64
	lastStatsTime    time.Time
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
	return container.NewVBox(
		title,
		widget.NewSeparator(),
		statusRow,
	)
}

// setConnectionStatus updates the connection label and the flag the dashboard
// reads to decide whether server stats should be polled over SSH.
func setConnectionStatus(connected bool) {
	tunnelConnected.Store(connected)
	if connectionStatus != nil {
		if connected {
			connectionStatus.SetText("Connection: Connected")
		} else {
			connectionStatus.SetText("Connection: Disconnected")
		}
	}
}

// RefreshDashboard updates the dashboard labels with real data.
func RefreshDashboard() {
	// Peer count comes from the local device list
	if devices, err := data.ReadDevicesFile(); err == nil {
		fyne.Do(func() {
			if peersLabel != nil {
				peersLabel.SetText(fmt.Sprintf("Peers: %d", len(devices)))
			}
		})
	}

	// Server statistics are queried in the background so the UI stays responsive.
	// Skip the SSH query while disconnected to avoid polling the server needlessly.
	go func() {
		if !tunnelConnected.Load() {
			fyne.Do(func() {
				if handshakeLabel != nil {
					handshakeLabel.SetText("Last Handshake: N/A")
				}
			})
			return
		}

		stats, err := wireguard.GetServerStats()
		if err != nil {
			// Log the failure once so the Logs tab shows why stats stay empty
			if !statsErrorLogged {
				statsErrorLogged = true
				writeLog("WARN", "Failed to read server stats: "+err.Error())
			}
			fyne.Do(func() {
				if handshakeLabel != nil {
					handshakeLabel.SetText("Last Handshake: Server unreachable")
				}
			})
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
