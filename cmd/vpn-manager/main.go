package main

import (
	"fyne.io/fyne/v2/app"

	"vpn-manager/internal/ui"
)

func main() {
	// Create a new Fyne application with a unique ID
	a := app.NewWithID("com.example.vpnmanager")

	w := a.NewWindow("VPN Manager")

	// Build the navigation and set it as the content of the window.
	ui.BuildWindow(a, w)

	w.ShowAndRun()
}
