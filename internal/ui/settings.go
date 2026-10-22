package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildSettings() fyne.CanvasObject {
	title := widget.NewLabel("Settings")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 1. Server settings form
	serverAddress := widget.NewEntry()
	sshPort := widget.NewEntry()
	sshUsername := widget.NewEntry()
	sshKeyPath := widget.NewEntry()
	connectionTimeout := widget.NewEntry()

	serverForm := widget.NewForm(
		widget.NewFormItem("Server Address", serverAddress),
		widget.NewFormItem("SSH Port", sshPort),
		widget.NewFormItem("SSH Username", sshUsername),
		widget.NewFormItem("SSH Key Path", sshKeyPath),
		widget.NewFormItem("Timeout (s)", connectionTimeout),
	)
	serverCard := widget.NewCard("Server Connection", "", serverForm)

	// 2. WireGuard settings form
	wireguardInterface := widget.NewEntry()
	refreshInterval := widget.NewEntry()

	wgForm := widget.NewForm(
		widget.NewFormItem("Interface", wireguardInterface),
		widget.NewFormItem("Refresh Interval", refreshInterval),
	)
	wgCard := widget.NewCard("WireGuard", "", wgForm)

	// 3. Appearance settings form
	closeToTray := widget.NewCheck("", func(b bool) {})
	autoRefresh := widget.NewCheck("", func(b bool) {})
	startMinimized := widget.NewCheck("", func(b bool) {})

	appearanceForm := widget.NewForm(
		widget.NewFormItem("Close to Tray", closeToTray),
		widget.NewFormItem("Auto Refresh", autoRefresh),
		widget.NewFormItem("Start Minimized", startMinimized),
	)
	appearanceCard := widget.NewCard("Appearance", "", appearanceForm)

	testConnectionButton := widget.NewButton(
		"Test Connection",
		func() {

		},
	)
	testConnectionButton.Importance = widget.SuccessImportance

	saveSettingsButton := widget.NewButton(
		"Save Settings",
		func() {

		},
	)
	saveSettingsButton.Importance = widget.HighImportance

	buttonBox := container.NewHBox(
		testConnectionButton,
		saveSettingsButton,
	)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		serverCard,
		wgCard,
		appearanceCard,
		widget.NewSeparator(),
		buttonBox,
	)

	settings := container.NewVScroll(content)

	return settings
}
