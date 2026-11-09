# vpn-manager

A Windows GUI application for managing a remote WireGuard server over SSH. It
sets up the server, creates client devices, generates their configuration files,
and brings client tunnels up and down on the local machine.

## Features

- Connect to a remote WireGuard server over SSH using key authentication.
- Generate and register SSH keys on the server.
- One-click server setup from the app: installs WireGuard, creates the `wg0`
  interface, and configures passwordless `sudo` for `wg` / `wg-quick`.
- Create and delete client devices (generates key pairs, registers peers on the
  server, and writes local client configs).
- Connect and disconnect client tunnels through the WireGuard Windows tunnel
  service.
- Dashboard with connection status, peer count, last handshake, download/upload
  totals, and total traffic. Stats are only polled while the tunnel is connected.
- Logs tab and a system tray menu.
- `scripts/setup.sh`: a one-touch firewall/NAT script for Fedora/RHEL servers.

## How it works

```
 Windows client (vpn-manager.exe)                    Linux server
 ┌──────────────────────────────┐                ┌────────────────────────┐
 │ GUI ── SSH (key auth) ───────┼── 192.168.x ──►│ sshd                   │
 │  │      (setup / devices)    │   (tunnel off) │                        │
 │  └─ WireGuardNT tunnel ──────┼── UDP 51820 ──►│ wg0 (10.10.10.1/24)    │
 │      (service + wireguard.dll)                │  └─ NAT ──► Internet   │
 │ Stats: SSH to 10.10.10.1 ────┼─ via tunnel ──►│ sshd (wireguard zone)  │
 └──────────────────────────────┘                └────────────────────────┘
```

While the tunnel is **connected**, the dashboard reads server stats over SSH to
the VPN gateway address (the `.1` of your VPN subnet, e.g. `10.10.10.1`). This is
required for full tunnels, where WireGuard's firewall blocks the server's LAN
address.

## Prerequisites

**Client (Windows)**

- Windows 10/11 (amd64), run as Administrator (see `app.manifest`).
- Go (version from `go.mod`) to build from source.
- `wireguard.dll` (WireGuardNT user-mode library) placed **next to
  `vpn-manager.exe`** (or on the system `PATH`). It is not committed to this
  repository (`*.dll` is git-ignored). Get it from the official WireGuardNT
  distribution: <https://www.wireguard.com/install/> (source:
  <https://github.com/WireGuard/wireguard-nt>). Use the `amd64` build.
- OpenSSH client (`ssh`, `ssh-keygen`, `ssh-keyscan`) — included with Windows,
  or with Git for Windows.

**Server (Linux)**

- A Linux server reachable over SSH with `sudo` available.
- `scripts/setup.sh` targets **Fedora/RHEL with firewalld**. On other
  distributions, apply the equivalent forwarding/NAT rules manually.
- UDP port `51820` reachable from the client (open it in your cloud/router
  firewall as well).

## Build

```powershell
git clone https://github.com/net-learner-04/vpn-manager.git
cd vpn-manager
go build .
```

Copy `wireguard.dll` next to the resulting `vpn-manager.exe`.

## Quick start

### 1. Create an SSH key on the client

```powershell
ssh-keygen -t ed25519 -f $env:USERPROFILE\.ssh\vpn-manager
```

This creates `vpn-manager` (private key) and `vpn-manager.pub` (public key).
Press Enter for an empty passphrase if you want the app to connect unattended.

### 2. Register the public key on the server

Either use **Register SSH Key** in the app's Settings tab, or do it manually:

```powershell
type $env:USERPROFILE\.ssh\vpn-manager.pub | ssh <user>@<server> "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

Verify key login works:

```powershell
ssh -i $env:USERPROFILE\.ssh\vpn-manager <user>@<server> echo OK
```

### 3. Prepare the server (firewall, forwarding, NAT)

On the server (Fedora/RHEL):

```bash
git clone https://github.com/net-learner-04/vpn-manager.git
cd vpn-manager
sudo bash scripts/setup.sh
```

The script is idempotent. It installs `wireguard-tools`, enables IPv4
forwarding, creates `wg0` if it does not exist, creates the firewalld
`wireguard` zone (SSH allowed), opens `51820/udp`, and enables masquerade plus a
forwarding policy for full-tunnel clients. Override defaults with environment
variables:

```bash
sudo VPN_INTERFACE=wg0 WG_PORT=51820 VPN_SERVER_IP=10.10.10.1/24 bash scripts/setup.sh
```

### 4. Configure the app

Run `vpn-manager.exe` as Administrator and open **Settings**:

| Setting            | Example                          | Notes                                       |
| ------------------ | -------------------------------- | ------------------------------------------- |
| Server address     | `203.0.113.10`                   | Server IP or hostname used for SSH/endpoint |
| SSH port / user    | `22` / `youruser`                |                                             |
| SSH key path       | `C:\Users\you\.ssh\vpn-manager`  | Private key from step 1                     |
| Interface / port   | `wg0` / `51820`                  |                                             |
| Subnet             | `10.10.10.0/24`                  | The server is `.1`                          |
| DNS                | `1.1.1.1`                        |                                             |
| Allowed IPs        | `0.0.0.0/0`                      | Full tunnel (see Notes)                     |
| Refresh interval   | `5`                              | Seconds between dashboard stat polls        |

Click **Test SSH Connection** to confirm. If you did not run `setup.sh`, click
**Setup Server** to install WireGuard and create `wg0` from the app (you still
need the firewall/NAT step from `setup.sh`).

### 5. Trust the server's tunnel address (one time)

The dashboard queries stats through the tunnel at `10.10.10.1`, which is a
different host key entry than your server's public address. Connect the tunnel
once (step 7), then add the host key **while the tunnel is up**:

```bash
# Git Bash (recommended)
ssh-keyscan -T 5 10.10.10.1 >> ~/.ssh/known_hosts
```

```powershell
# PowerShell: avoid ">>" in Windows PowerShell 5.1 (it writes UTF-16)
ssh-keyscan -T 5 10.10.10.1 | Out-File -Append -Encoding ascii $env:USERPROFILE\.ssh\known_hosts
```

### 6. Add a device

On the **Devices** tab, add a device (e.g. `dv1`). The app generates a key pair,
registers the peer on the server, and writes `conf\dv1.conf`.

### 7. Connect

Select the device and click **Connect**. Then verify:

```powershell
ping 10.10.10.1
curl ifconfig.me      # full tunnel: prints the server's public IP
```

On the server:

```bash
sudo wg show wg0      # "latest handshake" should be recent
```

## Notes

- **Split vs. full tunnel.** `AllowedIPs = 10.10.10.0/24` routes only the VPN
  subnet through the tunnel, so the dashboard traffic counters stay near zero
  unless you talk to `10.10.10.x`. `AllowedIPs = 0.0.0.0/0` sends all traffic
  through the server (needs forwarding and NAT — handled by `scripts/setup.sh`).
- **Allowed IPs apply to new devices only.** After changing it in Settings,
  delete and re-add the device so its `.conf` is regenerated.
- **Handshake updates.** WireGuard re-handshakes roughly every 2 minutes, so
  "Last Handshake" changing slowly is normal.
- **Refresh interval.** Each refresh runs one SSH command on the server. Use
  5 seconds or more.

## Troubleshooting

| Symptom | Likely cause / fix |
| ------- | ------------------ |
| Tunnel fails to start, `wireguard.dll` error | Put the `amd64` `wireguard.dll` next to `vpn-manager.exe`. |
| Tunnel service dies with `ERROR_NO_SUCH_GROUP` in the ring log | Fixed in recent versions (service SID is now enabled). Rebuild from `main`. |
| Dashboard shows "Server unreachable", log says `i/o timeout` to `10.10.10.1:22` | The server's `wireguard` firewalld zone does not allow SSH: `sudo firewall-cmd --permanent --zone=wireguard --add-service=ssh && sudo firewall-cmd --reload` |
| Log says `Unknown host key for 10.10.10.1` | Run step 5 while the tunnel is connected. |
| `connectex ... forbidden` when SSHing to the server's LAN/public address while connected | Expected with a full tunnel (WireGuard's firewall blocks it). SSH to `10.10.10.1` instead. |
| Connected, but no internet | IP forwarding/NAT is missing on the server. Re-run `scripts/setup.sh`. |
| No handshake | Check `51820/udp` is open on the server, your router/cloud firewall, and that the peer is listed in `sudo wg show wg0`. |

## Project layout

```
main.go                  Entry point (GUI mode and /tunnelservice mode)
assets/                  Embedded application icon
scripts/setup.sh         One-touch server firewall/NAT setup (Fedora/RHEL)
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
