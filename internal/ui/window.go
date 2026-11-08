package ui

import (
	"vpn-manager/assets"

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

		// Always hide to tray when the window is closed
		w.SetCloseIntercept(func() {
			w.Hide()
		})
	}

	navigation := BuildNavigation(w)

	w.SetContent(navigation)
}
