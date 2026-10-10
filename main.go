package main

import (
	"fyne.io/fyne/v2/app"

	"vpn-manager/internal/ui"
)

func main() {
	// Create a new Fyne application.
	a := app.New()

	w := a.NewWindow("VPN Manager")

	// Build the navigation and set it as the content of the window.
	ui.BuildWindow(a, w)

	w.ShowAndRun()
}
