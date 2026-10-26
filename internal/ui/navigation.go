package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildNavigation(w fyne.Window) fyne.CanvasObject {
	// Create the different views for the application.
	dashboardView := BuildDashboard()
	deviceView := BuildDevices(w)
	logsView := BuildLogs()
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

	split := container.NewBorder(
		nil,
		nil,
		menuWrapper,
		nil,
		contentArea,
	)

	return split
}
