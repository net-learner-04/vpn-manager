package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vpn-manager/internal/data"
	"vpn-manager/internal/wireguard"
)

func BuildNavigation(w fyne.Window) fyne.CanvasObject {
	// Create the different views for the application.
	dashboardView := BuildDashboard()

	// Keep the dashboard labels up to date
	StartDashboardRefresh()
	deviceView := BuildDevices(w)
	logsView := BuildLogs(w)
	settingsView := BuildSettings(w)

	// Create a tab container to hold the different views.
	contentArea := container.NewPadded(dashboardView)

	menuItems := []string{
		"Dashboard",
		"Devices",
		"Logs",
		"Settings",
	}

	menu := widget.NewList(
		// Return the number of items in the list.
		func() int { return len(menuItems) },
		// Create a new label for each item in the list.
		func() fyne.CanvasObject { return widget.NewLabel("Menu") },
		// Update the label for each item in the list.
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(menuItems[id])
		},
	)

	// Device selector for connecting/disconnecting a specific tunnel
	deviceSelect := widget.NewSelect([]string{}, func(string) {})
	deviceSelect.PlaceHolder = "Select device"

	loadDeviceOptions := func() {
		devices, err := data.ReadDevicesFile()
		if err != nil {
			devices = []data.Device{}
		}

		names := make([]string, 0, len(devices))
		for _, d := range devices {
			names = append(names, d.Name)
		}

		// Clear the selection when the selected device no longer exists
		stillExists := false
		for _, n := range names {
			if n == deviceSelect.Selected {
				stillExists = true
				break
			}
		}
		if !stillExists {
			deviceSelect.Selected = ""
		}

		deviceSelect.Options = names
		deviceSelect.Refresh()
	}
	loadDeviceOptions()

	connectButton := widget.NewButton(
		"Connect",
		func() {
			loadDeviceOptions()

			name := deviceSelect.Selected
			if name == "" {
				writeLog("WARN", "No device selected")
				showError(w, fmt.Errorf("Select a device to connect"))
				return
			}

			go func() {
				err := wireguard.ConnectWireguard(data.ConfPath(name))

				fyne.Do(func() {
					if err != nil {
						handleError(w, "Connect failed", err)
						return
					}
					writeLog("INFO", "Successfully connected to WireGuard.")
					connectionStatus.SetText("Connection: Connected")
				})
			}()
		},
	)
	connectButton.Importance = widget.HighImportance

	disconnectButton := widget.NewButton(
		"Disconnect",
		func() {
			loadDeviceOptions()

			name := deviceSelect.Selected
			if name == "" {
				writeLog("WARN", "No device selected")
				showError(w, fmt.Errorf("Select a device to disconnect"))
				return
			}

			go func() {
				err := wireguard.DisconnectWireguard(data.ConfPath(name))

				fyne.Do(func() {
					if err != nil {
						handleError(w, "Disconnect failed", err)
						return
					}
					writeLog("INFO", "Successfully disconnected from WireGuard.")
					connectionStatus.SetText("Connection: Disconnected")
				})
			}()
		},
	)
	disconnectButton.Importance = widget.DangerImportance

	// Buttons sit at the bottom of the menu, one per row
	footer := container.NewVBox(
		widget.NewSeparator(),
		deviceSelect,
		connectButton,
		disconnectButton,
	)

	menu.OnSelected = func(id widget.ListItemID) {
		switch menuItems[id] {
		case "Dashboard":
			contentArea.Objects = []fyne.CanvasObject{dashboardView}
		case "Devices":
			contentArea.Objects = []fyne.CanvasObject{deviceView}
		case "Logs":
			contentArea.Objects = []fyne.CanvasObject{logsView}
		case "Settings":
			contentArea.Objects = []fyne.CanvasObject{settingsView}
		}
		// Refresh the content area to display the selected view.
		contentArea.Refresh()
	}

	// Display with the Dashboard selected when first launched
	menu.Select(0)

	// A wrapper to fix the menu width
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(125, 0))
	menuWrapper := container.NewStack(spacer, menu)

	divider := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	divider.SetMinSize(fyne.NewSize(1, 0))

	// Menu on top, buttons pinned to the bottom of the left panel
	left := container.NewBorder(nil, footer, nil, divider, menuWrapper)

	split := container.NewBorder(
		nil,
		nil,
		left,
		nil,
		contentArea,
	)

	return split
}
