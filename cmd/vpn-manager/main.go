package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"vpn-manager/internal/ui"

	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

func main() {
	// Service mode: run the WireGuard tunnel service for the given config.
	if len(os.Args) > 1 && os.Args[1] == "/tunnelservice" {
		if len(os.Args) < 3 {
			os.Exit(1)
		}

		confPath := os.Args[2]
		conf.PresetRootDirectory(filepath.Dir(confPath))

		if err := tunnel.Run(confPath); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	// Create a new Fyne application with a unique ID
	a := app.NewWithID("com.example.vpnmanager")

	w := a.NewWindow("VPN Manager")

	// Build the navigation and set it as the content of the window.
	ui.BuildWindow(a, w)

	w.ShowAndRun()
}
