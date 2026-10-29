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

	knownHostsPath := widget.NewEntry()

	connectionTimeout := widget.NewEntry()
	connectionTimeout.SetPlaceHolder("e.g. 10")

	// WireGuard settings form
	wireguardInterface := widget.NewEntry()
	refreshInterval := widget.NewEntry()

	// Appearance settings form
	closeToTray := widget.NewCheck("", func(b bool) {})

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
		knownHostsPath.SetText(config.SSHKnownHostsPath)
		connectionTimeout.SetText(config.ConnectionTimeout)

		wireguardInterface.SetText(config.WireGuardInterface)
		refreshInterval.SetText(config.RefreshInterval)

		closeToTray.SetChecked(config.CloseToTray)
	}

	// Load data when the program starts
	loadConfig()

	serverForm := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewForm(
				widget.NewFormItem("Server Address", serverAddress),
				widget.NewFormItem("SSH Username", sshUsername),
			),
			widget.NewForm(
				widget.NewFormItem("SSH Port", sshPort),
				widget.NewFormItem("Timeout (s)", connectionTimeout),
			),
		),
		widget.NewForm(
			widget.NewFormItem("SSH Key Path", sshKeyPath),
			widget.NewFormItem("Known Hosts Path", knownHostsPath),
		),
	)
	serverCard := widget.NewCard("Server Connection", "", serverForm)

	wgForm := widget.NewForm(
		widget.NewFormItem("Interface", wireguardInterface),
		widget.NewFormItem("Refresh Interval", refreshInterval),
	)
	wgCard := widget.NewCard("WireGuard", "", wgForm)

	appearanceForm := widget.NewForm(
		widget.NewFormItem("Close to Tray", closeToTray),
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
				SSHKnownHostsPath: knownHostsPath.Text,
				ConnectionTimeout: connectionTimeout.Text,

				WireGuardInterface: wireguardInterface.Text,
				RefreshInterval:    refreshInterval.Text,

				CloseToTray: closeToTray.Checked,
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
		container.NewGridWithColumns(2, wgCard, appearanceCard),
		widget.NewSeparator(),
		buttonBox,
	)

	return content
}
