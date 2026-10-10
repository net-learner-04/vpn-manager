package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func BuildLogs() fyne.CanvasObject {
	return widget.NewLabel("Logs Screen")
}
