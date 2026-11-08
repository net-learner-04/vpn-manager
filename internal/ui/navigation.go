package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vpn-manager/internal/wireguard"
)

func BuildNavigation(w fyne.Window) fyne.CanvasObject {
	// Create the different views for the application.
	dashboardView := BuildDashboard()
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

	connectButton := widget.NewButton(
		"Connect",
		func() {
			err := wireguard.ConnectWireguard()
			if err != nil {
				writeLog("ERROR", err.Error())
				return
			}
			writeLog("INFO", "Successfully connected to WireGuard.")
			connectionStatus.SetText("Connection: Connected")
		},
	)
	connectButton.Importance = widget.HighImportance

	disconnectButton := widget.NewButton(
		"Disconnect",
		func() {
			err := wireguard.DisconnectWireguard()
			if err != nil {
				writeLog("ERROR", err.Error())
				return
			}
			writeLog("INFO", "Successfully disconnected from WireGuard.")
			connectionStatus.SetText("Connection: Disconnected")
		},
	)
	disconnectButton.Importance = widget.DangerImportance

	// Buttons sit at the bottom of the menu, one per row
	footer := container.NewVBox(
		widget.NewSeparator(),
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
