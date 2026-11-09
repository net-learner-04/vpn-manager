package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/app"

	"vpn-manager/internal/ui"
	"vpn-manager/internal/wireguard"

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

	// Warn when running from a temporary "go run" build directory, because the
	// service binary path would be unstable and lost on the next run.
	if exe, err := os.Executable(); err == nil {
		lower := strings.ToLower(exe)
		if strings.Contains(lower, "go-build") || strings.Contains(lower, "\\temp\\") {
			log.Println("WARNING: running from a temporary build directory (go run). Use 'go build .' and run the resulting .exe instead.")
		}
	}

	// Clean up any orphan tunnel services left behind by an abnormal exit
	go func() {
		if err := wireguard.CleanupOrphanServices(); err != nil {
			log.Println("Cleanup orphan services:", err)
		}
	}()

	// Create a new Fyne application with a unique ID
	a := app.NewWithID("com.example.vpnmanager")

	w := a.NewWindow("VPN Manager")

	// Build the navigation and set it as the content of the window.
	ui.BuildWindow(a, w)

	w.ShowAndRun()
}
