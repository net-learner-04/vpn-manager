package ui

import (
	"vpn-manager/assets"
	"vpn-manager/internal/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

func BuildWindow(a fyne.App, w fyne.Window) {
	a.Settings().SetTheme(&Theme{})

	iconRes := fyne.NewStaticResource("icon.png", assets.Icon)

	a.SetIcon(iconRes)

	w.SetIcon(iconRes)
	w.Resize(fyne.NewSize(700, 450))
	w.SetFixedSize(true)

	// Load settings
	config, err := data.ReadSettingsFile()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	// Set up the system tray menu if the app supports it
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("VPN Manager",
			fyne.NewMenuItem("Open", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(iconRes)

		// Handle window closing based on settings
		w.SetCloseIntercept(func() {
			if config.CloseToTray {
				// Hide the window but keep the app running
				w.Hide()
			} else {
				// Exit the application
				a.Quit()
			}
		})
	}

	navigation := BuildNavigation(w)

	w.SetContent(navigation)
}
