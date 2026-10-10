package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildNavigation() fyne.CanvasObject {
	// Create the different views for the application.
	dashboardView := BuildDashboard()
	deviceView := BuildDevices()
	trafficView := BuildTraffic()
	logsView := BuildLogs()
	settingsView := BuildSettings()

	// Create a tab container to hold the different views.
	contentArea := container.NewPadded(dashboardView)

	menuItems := []string{
		"Dashboard",
		"Devices",
		"Traffic",
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

	menu.OnSelected = func(id widget.ListItemID) {
		switch menuItems[id] {
		case "Dashboard":
			contentArea.Objects = []fyne.CanvasObject{dashboardView}
		case "Devices":
			contentArea.Objects = []fyne.CanvasObject{deviceView}
		case "Traffic":
			contentArea.Objects = []fyne.CanvasObject{trafficView}
		case "Logs":
			contentArea.Objects = []fyne.CanvasObject{logsView}
		case "Settings":
			contentArea.Objects = []fyne.CanvasObject{settingsView}
		}
		// Refresh the content area to display the selected view.
		contentArea.Refresh()
	}

	split := container.NewHSplit(
		menu,
		contentArea,
	)

	// Set the initial position of the split to 20%
	// for the menu and 80% for the content area.
	split.SetOffset(0.2)

	return split
}
