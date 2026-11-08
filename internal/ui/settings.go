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
	serverInterface := widget.NewEntry()
	refreshInterval := widget.NewEntry()
	wireguardPort := widget.NewEntry()
	dns := widget.NewEntry()
	allowedIPs := widget.NewEntry()
	subnet := widget.NewEntry()

	loadConfig := func() {
		config, err := data.ReadSettingsFile()
		if err != nil {
			writeLog("ERROR", err.Error())
			dialog.ShowError(err, w)
			return
		}

		// Server Connection
		serverAddress.SetText(config.ServerAddress)
		sshPort.SetText(config.SSHPort)
		sshUsername.SetText(config.SSHUsername)
		sshKeyPath.SetText(config.SSHKeyPath)
		knownHostsPath.SetText(config.SSHKnownHostsPath)
		connectionTimeout.SetText(config.ConnectionTimeout)

		// WireGuard
		serverInterface.SetText(config.ServerInterface)
		refreshInterval.SetText(config.RefreshInterval)
		wireguardPort.SetText(config.WireGuardPort)
		dns.SetText(config.DNS)
		allowedIPs.SetText(config.AllowedIPs)
		subnet.SetText(config.VPNSubnet)
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

	wgForm := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewForm(
				widget.NewFormItem("Interface", serverInterface),
				widget.NewFormItem("VPN Subnet", subnet),
				widget.NewFormItem("DNS", dns),
			),
			widget.NewForm(
				widget.NewFormItem("WireGuard Port", wireguardPort),
				widget.NewFormItem("Allowed IPs", allowedIPs),
				widget.NewFormItem("Refresh (s)", refreshInterval),
			),
		),
	)
	wgCard := widget.NewCard("WireGuard", "", wgForm)

	// Test Connection button
	var testConnectionButton *widget.Button
	testConnectionButton = widget.NewButton(
		"Test SSH Connection",
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
			privPath := sshKeyPath.Text
			pubPath := fmt.Sprintf("%s.pub", privPath)

			if privPath == "" {
				dialog.ShowError(fmt.Errorf("SSH Key Path is empty"), w)
				return
			}

			// Ask for the server password and register the key
			askPassword := func() {
				password := widget.NewPasswordEntry()
				items := []*widget.FormItem{
					widget.NewFormItem("Server Password", password),
				}

				dialog.ShowForm("Register SSH Key", "Register", "Cancel", items, func(ok bool) {
					if !ok {
						return
					}

					pw := password.Text
					password.SetText("")

					// Run in a goroutine so the UI thread is not blocked
					go func() {
						err := sshclient.RegisterKey(pw, privPath, pubPath)

						// With Fyne 2.6+, wrap UI calls in fyne.Do(func() { ... })
						if err != nil {
							writeLog("ERROR", "Failed to register SSH key: "+err.Error())
							dialog.ShowError(err, w)
							return
						}

						writeLog("INFO", "SSH public key registered successfully.")
						dialog.ShowInformation(
							"SSH Key Registered",
							"SSH public key registered successfully.",
							w,
						)
					}()
				}, w)
			}

			// If no key file exists, go straight to the password prompt
			if _, err := os.Stat(privPath); err != nil {
				askPassword()
				return
			}

			// Try the existing key first
			go func() {
				client, err := sshclient.ConnectServer()
				if err == nil {
					client.Close()
					writeLog("INFO", "SSH key is already registered.")
					dialog.ShowInformation("Already Registered", "SSH key login already works.", w)
					return
				}

				// Only ask for a password when the key was rejected
				if sshclient.IsAuthError(err) {
					askPassword()
					return
				}

				// Network, known_hosts, or other errors: do not ask for a password
				writeLog("ERROR", "SSH connection failed: "+err.Error())
				dialog.ShowError(fmt.Errorf("SSH connection failed: %w", err), w)
			}()
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

				ServerInterface: serverInterface.Text,
				WireGuardPort:   wireguardPort.Text,
				DNS:             dns.Text,
				AllowedIPs:      allowedIPs.Text,
				RefreshInterval: refreshInterval.Text,
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
		wgCard,
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
