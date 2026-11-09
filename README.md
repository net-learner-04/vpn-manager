# vpn-manager

A Windows GUI application for managing a remote WireGuard server over SSH. It
sets up the server, creates client devices, generates their configuration files,
and brings client tunnels up and down on the local machine.

## Features

- Connect to a remote WireGuard server over SSH using key authentication.
- Generate and register SSH keys on the server.
- One-click server setup: installs WireGuard, creates the `wg0` interface, and
  configures passwordless `sudo` for `wg` / `wg-quick`.
- Create and delete client devices (generates key pairs, registers peers on the
  server, and writes local client configs).
- Connect and disconnect client tunnels through the WireGuard Windows tunnel
  service.
- Dashboard with connection status, peer count, last handshake, download/upload
  rates, and total traffic.
- Logs tab and a system tray menu.

## Prerequisites

- Windows (amd64).
- Go 1.27 or newer.
- A Linux WireGuard server reachable over SSH, with `sudo` available.
- The WireGuardNT driver and its user-mode `wireguard.dll`. These ship with
  [WireGuard for Windows](https://www.wireguard.com/install/); the driver source
  is at <https://github.com/WireGuard/wireguard-nt>. `wireguard.dll` must be
  resolvable by the app, for example placed next to `vpn-manager.exe` or on the
  system `PATH`.

The application requests administrator privileges (see `app.manifest`).

## Build & run

```powershell
go build .
.\vpn-manager.exe
```

Run the resulting executable as Administrator.

## Usage

1. On the **Settings** tab, enter the server address, SSH port, username, SSH key
   path, and WireGuard settings (interface, port, DNS, allowed IPs, subnet).
2. Use **Generate SSH Key** / **Register SSH Key** (or **Test SSH Connection**) to
   set up key authentication with the server.
3. Click **Setup Server** to install WireGuard and create `wg0` on the server.
4. On the **Devices** tab, add a device.
5. Select the device and click **Connect** in the left panel.

## Project layout

```
main.go                  Entry point (GUI mode and /tunnelservice mode)
assets/                  Embedded application icon
internal/data/           Settings/devices storage and paths (ProgramData)
internal/log/            File logger
internal/ssh/            SSH client, key generation, and key registration
internal/ui/             Fyne UI (dashboard, devices, logs, settings, navigation)
internal/wireguard/      Server setup, device/peer management, stats, and
                         connect/disconnect through the Windows tunnel service
```

Runtime data is stored under `%ProgramData%\VPNManager`:

- `settings.json` — server and WireGuard settings
- `devices.json` — managed devices
- `conf/<name>.conf` — generated client configuration files
- `logs/vpn-manager.log` — application log

## Notes

- A split tunnel (e.g. `AllowedIPs = 10.10.10.0/24`) routes only the VPN subnet
  through the tunnel. Use `AllowedIPs = 0.0.0.0/0` for a full tunnel; the server
  must then enable IP forwarding and NAT so client traffic can egress through it.

