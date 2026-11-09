package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"vpn-manager/internal/data"
	"vpn-manager/internal/log"
	sshclient "vpn-manager/internal/ssh"
	"vpn-manager/internal/wireguard"

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

// handleError logs an error and shows it in a dialog.
func handleError(w fyne.Window, message string, err error) {
	writeLog("ERROR", message+": "+err.Error())
	showError(w, err)
}

// showError displays an error in a dialog that wraps long messages instead of
// stretching the window.
func showError(w fyne.Window, err error) {
	message := err.Error()
	if len(message) > 800 {
		message = message[:800] + "..."
	}

	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(label)

	// Size the dialog to the message: compact for short text, wider for long text
	width := float32(380)
	height := float32(160)
	if len(message) > 100 {
		width = 560
		height = 240
	}

	d := dialog.NewCustom("Error", "OK", scroll, w)
	d.Show()
	d.Resize(fyne.NewSize(width, height))
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
	allowedIPs.SetPlaceHolder("e.g. 0.0.0.0/0 (all traffic)")
	subnet := widget.NewEntry()
	subnet.SetPlaceHolder("e.g. 10.10.10.0/24")

	loadConfig := func() {
		config, err := data.ReadSettingsFile()
		if err != nil {
			handleError(w, "Failed to load settings", err)
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

	// Autosave settings on every change so the app always uses the latest values
	serverPublicKey := ""
	if config, err := data.ReadSettingsFile(); err == nil {
		serverPublicKey = config.ServerPublicKey
	}
	originalServerAddress := serverAddress.Text

	saveConfig := func() {
		if serverAddress.Text != originalServerAddress {
			serverPublicKey = ""
		}

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
			VPNSubnet:       subnet.Text,
			ServerPublicKey: serverPublicKey,
		}

		if err := data.WriteSettingsFile(config); err != nil {
			writeLog("ERROR", "Failed to autosave settings: "+err.Error())
		}
	}

	serverAddress.OnChanged = func(string) { saveConfig() }
	sshPort.OnChanged = func(string) { saveConfig() }
	sshUsername.OnChanged = func(string) { saveConfig() }
	sshKeyPath.OnChanged = func(string) { saveConfig() }
	knownHostsPath.OnChanged = func(string) { saveConfig() }
	connectionTimeout.OnChanged = func(string) { saveConfig() }
	serverInterface.OnChanged = func(string) { saveConfig() }
	refreshInterval.OnChanged = func(string) { saveConfig() }
	wireguardPort.OnChanged = func(string) { saveConfig() }
	dns.OnChanged = func(string) { saveConfig() }
	allowedIPs.OnChanged = func(string) { saveConfig() }
	subnet.OnChanged = func(string) { saveConfig() }

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

	var testConnection func()
	testConnection = func() {
		testConnectionButton.Disable()

		go func() {
			// Connect to the specified server
			client, err := sshclient.ConnectServer()

			if err != nil {
				var hostKeyErr *sshclient.HostKeyError
				if errors.As(err, &hostKeyErr) {
					// First contact: ask the user to trust the server key
					fyne.Do(func() {
						dialog.NewConfirm(
							"Unknown Server",
							"First connection to this server.\nFingerprint: "+hostKeyErr.Fingerprint+"\n\nTrust this server?",
							func(trust bool) {
								if !trust {
									testConnectionButton.Enable()
									return
								}

								if err := sshclient.TrustHostKey(hostKeyErr.Address, hostKeyErr.Key); err != nil {
									handleError(w, "Failed to save host key", err)
									testConnectionButton.Enable()
									return
								}

								// Retry now that the key is trusted
								fyne.Do(testConnection)
							},
							w,
						).Show()
					})
					return
				}

				fyne.Do(func() {
					handleError(w, "SSH connection failed", err)
					testConnectionButton.Enable()
				})
				return
			}

			// Close the SSH connection after testing
			err = client.Close()

			fyne.Do(func() {
				if err != nil {
					handleError(w, "SSH connected, but closing failed", err)
				} else {
					writeLog("INFO", "Successfully connected to the SSH server.")
					dialog.ShowInformation("Connection Test", "Successfully connected to the SSH server.", w)
				}

				testConnectionButton.Enable()
			})
		}()
	}

	testConnectionButton = widget.NewButton("Test SSH Connection", testConnection)

	// Generate SSH Key button
	generateSSHKeyButton := widget.NewButton(
		"Generate SSH Key",
		func() {
			privateKey, publicKey, err := sshclient.GenerateSSHKey()
			if err != nil {
				handleError(w, "Failed to generate SSH key", err)
				return
			}

			// Save the generated keys to the specified paths
			if err := sshclient.SavePrivateKey(privateKey, sshKeyPath.Text); err != nil {
				handleError(w, "Failed to save SSH private key", err)
				return
			}

			// Save the public key with a .pub extension
			if err := sshclient.SavePublicKey(
				publicKey,
				fmt.Sprintf("%s.pub", sshKeyPath.Text),
			); err != nil {
				handleError(w, "Failed to save SSH public key", err)
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
				showError(w, fmt.Errorf("SSH Key Path is empty"))
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

						fyne.Do(func() {
							if err != nil {
								handleError(w, "Failed to register SSH key", err)
								return
							}

							writeLog("INFO", "SSH public key registered successfully.")
							dialog.ShowInformation(
								"SSH Key Registered",
								"SSH public key registered successfully.",
								w,
							)
						})
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
					fyne.Do(func() {
						writeLog("INFO", "SSH key is already registered.")
						dialog.ShowInformation("Already Registered", "SSH key login already works.", w)
					})
					return
				}

				// Only ask for a password when the key was rejected
				if sshclient.IsAuthError(err) {
					fyne.Do(func() {
						askPassword()
					})
					return
				}

				// Network, known_hosts, or other errors: do not ask for a password
				fyne.Do(func() {
					handleError(w, "SSH connection failed", err)
				})
			}()
		},
	)

	// Setup Server button
	var setupServerButton *widget.Button
	setupServerButton = widget.NewButton(
		"Setup Server",
		func() {
			setupServerButton.Disable()

			go func() {
				err := wireguard.SetupServer()
				if err == nil {
					err = wireguard.SetupSudoers()
				}

				fyne.Do(func() {
					setupServerButton.Enable()
					if err != nil {
						handleError(w, "Server setup failed", err)
						return
					}
					writeLog("INFO", "Server setup completed successfully.")
					dialog.ShowInformation("Setup Server", "Server setup completed successfully.", w)
				})
			}()
		},
	)

	// Save Settings button
	saveSettingsButton := widget.NewButton(
		"Save Settings",
		func() {
			saveConfig()
			writeLog("INFO", "Settings saved successfully.")
			dialog.ShowInformation("Settings Saved", "Settings have been saved successfully.", w)
		},
	)

	buttonBox := container.NewGridWithColumns(
		3,
		testConnectionButton,
		generateSSHKeyButton,
		registerSSHKeyButton,
		setupServerButton,
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
