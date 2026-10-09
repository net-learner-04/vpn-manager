package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"

	"vpn-manager/internal/ui"
)

func main() {
	fmt.Println("Starting VPN Manager...")

	// Create a new Fyne application.
	a := app.New()
	w := a.NewWindow("VPN Manager")

	ui.BuildWindow(w)

	w.ShowAndRun()

	fmt.Println("VPN Manager exited.")
}
