package ui

import (
	"vpn-manager/internal/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func BuildSettings(w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Settings")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Server settings form
	serverAddress := widget.NewEntry()
	serverAddress.SetPlaceHolder("e.g. 192.168.0.100")

	sshPort := widget.NewEntry()
	sshPort.SetPlaceHolder("e.g. 22")

	sshUsername := widget.NewEntry()
	sshUsername.SetPlaceHolder("e.g. root")

	sshKeyPath := widget.NewEntry()
	sshKeyPath.SetPlaceHolder("e.g. C:\\Users\\Username\\.ssh\\id_rsa")

	connectionTimeout := widget.NewEntry()
	connectionTimeout.SetPlaceHolder("e.g. 10")

	// WireGuard settings form
	wireguardInterface := widget.NewEntry()
	refreshInterval := widget.NewEntry()

	// Appearance settings form
	closeToTray := widget.NewCheck("", func(b bool) {})
	autoRefresh := widget.NewCheck("", func(b bool) {})

	loadConfig := func() {
		config, err := data.ReadSettingsFile()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		serverAddress.SetText(config.ServerAddress)
		sshPort.SetText(config.SSHPort)
		sshUsername.SetText(config.SSHUsername)
		sshKeyPath.SetText(config.SSHKeyPath)
		connectionTimeout.SetText(config.ConnectionTimeout)

		wireguardInterface.SetText(config.WireGuardInterface)
		refreshInterval.SetText(config.RefreshInterval)

		closeToTray.SetChecked(config.CloseToTray)
		autoRefresh.SetChecked(config.AutoRefresh)
	}

	// Load data when the program starts
	loadConfig()

	serverForm := widget.NewForm(
		widget.NewFormItem("Server Address", serverAddress),
		widget.NewFormItem("SSH Port", sshPort),
		widget.NewFormItem("SSH Username", sshUsername),
		widget.NewFormItem("SSH Key Path", sshKeyPath),
		widget.NewFormItem("Timeout (s)", connectionTimeout),
	)
	serverCard := widget.NewCard("Server Connection", "", serverForm)

	wgForm := widget.NewForm(
		widget.NewFormItem("Interface", wireguardInterface),
		widget.NewFormItem("Refresh Interval", refreshInterval),
	)
	wgCard := widget.NewCard("WireGuard", "", wgForm)

	appearanceForm := widget.NewForm(
		widget.NewFormItem("Close to Tray", closeToTray),
		widget.NewFormItem("Auto Refresh", autoRefresh),
	)
	appearanceCard := widget.NewCard("Appearance", "", appearanceForm)

	testConnectionButton := widget.NewButton(
		"Test Connection",
		func() {
			// to do test connection
		},
	)

	saveSettingsButton := widget.NewButton(
		"Save Settings",
		func() {
			config := data.Settings{
				ServerAddress:     serverAddress.Text,
				SSHPort:           sshPort.Text,
				SSHUsername:       sshUsername.Text,
				SSHKeyPath:        sshKeyPath.Text,
				ConnectionTimeout: connectionTimeout.Text,

				WireGuardInterface: wireguardInterface.Text,
				RefreshInterval:    refreshInterval.Text,

				CloseToTray: closeToTray.Checked,
				AutoRefresh: autoRefresh.Checked,
			}

			if err := data.WriteSettingsFile(config); err != nil {
				dialog.ShowError(err, w)
				return
			}

			dialog.ShowInformation("Settings", "Settings saved successfully.", w)
		},
	)

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
