package ui

import (
	"vpn-manager/assets"
	"vpn-manager/internal/data"
	"vpn-manager/internal/wireguard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func BuildWindow(a fyne.App, w fyne.Window) {
	a.Settings().SetTheme(&Theme{})

	iconRes := fyne.NewStaticResource("icon.png", assets.Icon)

	a.SetIcon(iconRes)

	w.SetIcon(iconRes)
	w.Resize(fyne.NewSize(700, 450))
	w.SetFixedSize(true)

	// Set up the system tray menu if the app supports it
	if desk, ok := a.(desktop.App); ok {
		connectItem := fyne.NewMenuItem("Connect", func() {
			devices, err := data.ReadDevicesFile()
			if err != nil || len(devices) == 0 {
				a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "No devices available"})
				return
			}

			go func() {
				err := wireguard.ConnectWireguard(data.ConfPath(devices[0].Name))
				fyne.Do(func() {
					if err != nil {
						a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "Connect failed: " + err.Error()})
						return
					}
					a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "Connected"})
				})
			}()
		})

		disconnectItem := fyne.NewMenuItem("Disconnect", func() {
			devices, err := data.ReadDevicesFile()
			if err != nil || len(devices) == 0 {
				a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "No devices available"})
				return
			}

			go func() {
				err := wireguard.DisconnectWireguard(data.ConfPath(devices[0].Name))
				fyne.Do(func() {
					if err != nil {
						a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "Disconnect failed: " + err.Error()})
						return
					}
					a.SendNotification(&fyne.Notification{Title: "VPN Manager", Content: "Disconnected"})
				})
			}()
		})

		menu := fyne.NewMenu("VPN Manager",
			fyne.NewMenuItem("Open", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItemSeparator(),
			connectItem,
			disconnectItem,
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(iconRes)

		// Always hide to tray when the window is closed
		w.SetCloseIntercept(func() {
			w.Hide()
		})
	}

	navigation := BuildNavigation(w)

	w.SetContent(navigation)
}
