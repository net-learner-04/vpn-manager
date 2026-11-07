package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"vpn-manager/internal/app"
	"vpn-manager/internal/data"
	"vpn-manager/internal/log"
	sshclient "vpn-manager/internal/ssh"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func writeLog(level string, message string) error {
	_, file, _, ok := runtime.Caller(1)
	source := "unknown"
	if ok {
		dir := filepath.Base(filepath.Dir(file))
		name := filepath.Base(file)
		source = dir + "/" + name
	}

	return log.WriteLog(time.Now().Format("15:04:05"), level, message, source)
}

func BuildSettings(w fyne.Window) fyne.CanvasObject {
	// Title label
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

	knownHostsPath := widget.NewEntry()

	connectionTimeout := widget.NewEntry()
	connectionTimeout.SetPlaceHolder("e.g. 10")

	// WireGuard settings form
	wireguardInterface := widget.NewEntry()
	refreshInterval := widget.NewEntry()

	// Appearance settings form
	closeToTray := widget.NewCheck("", func(b bool) {})
	startMinimized := widget.NewCheck("", func(b bool) {})

	loadConfig := func() {
		config, err := data.ReadSettingsFile()
		if err != nil {
			writeLog("ERROR", err.Error())
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
		startMinimized.SetChecked(config.StartMinimized)
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
		widget.NewFormItem("Start Minimized", startMinimized),
	)
	appearanceCard := widget.NewCard("Appearance", "", appearanceForm)

	// Test Connection button
	var testConnectionButton *widget.Button
	testConnectionButton = widget.NewButton(
		"Test Connection",
		func() {
			testConnectionButton.Disable()

			go func() {
				// Connect to the specified server
				client, err := sshclient.ConnectServer()

				if err != nil {
					writeLog("ERROR", "SSH connection failed: "+err.Error())
					fyne.Do(func() {
						dialog.ShowError(
							fmt.Errorf("SSH connection failed: %w", err),
							w,
						)
						testConnectionButton.Enable()
					})
					return
				}

				// Close the SSH connection after testing
				err = client.Close()

				fyne.Do(func() {
					if err != nil {
						writeLog("ERROR", "SSH connected, but closing failed: "+err.Error())
						dialog.ShowError(
							fmt.Errorf("SSH connected, but closing failed: %w", err),
							w,
						)
					} else {
						writeLog("INFO", "Successfully connected to the SSH server.")
						dialog.ShowInformation(
							"Connection Test",
							"Successfully connected to the SSH server.",
							w,
						)
					}

					testConnectionButton.Enable()
				})
			}()
		},
	)

	// Generate SSH Key button
	generateSSHKeyButton := widget.NewButton(
		"Generate SSH Key",
		func() {
			privateKey, publicKey, err := sshclient.GenerateSSHKey()
			if err != nil {
				writeLog("ERROR", "Failed to generate SSH key: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			// Save the generated keys to the specified paths
			if err := sshclient.SavePrivateKey(privateKey, sshKeyPath.Text); err != nil {
				writeLog("ERROR", "Failed to save SSH private key: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			// Save the public key with a .pub extension
			if err := sshclient.SavePublicKey(
				publicKey,
				fmt.Sprintf("%s.pub", sshKeyPath.Text),
			); err != nil {
				writeLog("ERROR", "Failed to save SSH public key: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			writeLog("INFO", "SSH key pair generated and saved successfully.")
			dialog.ShowInformation(
				"SSH Key Generated",
				"SSH key pair generated and saved successfully.",
				w,
			)
		},
	)

	// Register SSH Key button
	registerSSHKeyButton := widget.NewButton(
		"Register SSH Key",
		func() {
			// Connect to the specified server
			client, err := sshclient.ConnectServer()
			if err != nil {
				writeLog("ERROR", "SSH connection failed: "+err.Error())
				dialog.ShowError(
					fmt.Errorf("SSH connection failed: %w", err),
					w,
				)
				return
			}
			defer client.Close()

			// Read the public key file
			publicKeyData, err := os.ReadFile(
				fmt.Sprintf("%s.pub", sshKeyPath.Text),
			)
			if err != nil {
				writeLog("ERROR", "Failed to read SSH public key: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			// Send the public key to the remote server
			if err := sshclient.SendPublicKey(client, publicKeyData); err != nil {
				writeLog("ERROR", "Failed to register SSH public key: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			writeLog("INFO", "SSH public key registered successfully.")
			dialog.ShowInformation(
				"SSH Key Registered",
				"SSH public key registered successfully.",
				w,
			)
		},
	)

	// Save Settings button
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

				CloseToTray:    closeToTray.Checked,
				StartMinimized: startMinimized.Checked,
			}

			if err := data.WriteSettingsFile(config); err != nil {
				writeLog("ERROR", "Failed to save settings: "+err.Error())
				dialog.ShowError(err, w)
				return
			}

			writeLog("INFO", "Settings saved successfully.")

			restartDialog := dialog.NewConfirm(
				"Settings Saved",
				"Settings have been saved successfully.\nThe application will now restart.",
				func(confirmed bool) {
					if !confirmed {
						return
					}

					if err := app.RestartApp(); err != nil {
						log.WriteLog(time.Now().Format("15:04:05"), "ERROR", err.Error(), "Application")
						dialog.ShowError(err, w)
					}
				},
				w,
			)

			restartDialog.SetConfirmText("Restart")
			restartDialog.SetDismissText("Cancel")
			restartDialog.Show()
		},
	)

	buttonBox := container.NewGridWithColumns(
		4,
		testConnectionButton,
		generateSSHKeyButton,
		registerSSHKeyButton,
		saveSettingsButton,
	)

	topContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		serverCard,
		container.NewGridWithColumns(2, wgCard, appearanceCard),
	)

	bottomContent := container.NewVBox(
		widget.NewSeparator(),
		buttonBox,
	)

	content := container.NewBorder(
		topContent,
		bottomContent,
		nil,
		nil,
	)

	return content
}
