package ui

import (
	"vpn-manager/assets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func BuildWindow(a fyne.App, w fyne.Window) {
	iconRes := fyne.NewStaticResource("icon.png", assets.Icon)

	a.SetIcon(iconRes)

	w.SetIcon(iconRes)
	w.Resize(fyne.NewSize(600, 400))
	w.SetFixedSize(true)

	// Set up the system tray menu if the app supports it.
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("VPN Manager",
			fyne.NewMenuItem("Open", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		)
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(iconRes)

		// Set the close intercept to hide
		// the window instead of closing it.
		w.SetCloseIntercept(func() { w.Hide() })
	}

	navigation := BuildNavigation()

	w.SetContent(navigation)
}
