#!/usr/bin/env bash
set -euo pipefail

# vpn-manager: one-touch server setup for Fedora/RHEL (firewalld).
#
# What it does (idempotent, safe to re-run):
#   1. Installs wireguard-tools
#   2. Enables IPv4 forwarding (persistent)
#   3. Creates /etc/wireguard/<iface>.conf if missing and enables wg-quick@<iface>
#   4. Creates the "wireguard" firewalld zone, attaches the VPN interface,
#      and allows SSH inside it (the manager app queries stats over the tunnel)
#   5. Opens the WireGuard UDP port on the WAN zone
#   6. Enables masquerade (NAT) on the WAN zone + a forwarding policy
#      so full-tunnel clients (AllowedIPs = 0.0.0.0/0) can reach the internet
#
# Usage:
#     sudo bash setup.sh
#
# Optional overrides:
#     sudo VPN_INTERFACE=wg0 WG_PORT=51820 VPN_SERVER_IP=10.10.10.1/24 bash setup.sh

VPN_INTERFACE="${VPN_INTERFACE:-wg0}"
WG_PORT="${WG_PORT:-51820}"
VPN_SERVER_IP="${VPN_SERVER_IP:-10.10.10.1/24}"
POLICY_NAME="vpn-egress"

log()  { echo -e "\033[1;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*" >&2; }
die()  { echo -e "\033[1;31m[x]\033[0m $*" >&2; exit 1; }

# --- Preflight ---------------------------------------------------------------
[ "$(id -u)" -eq 0 ] || die "Run as root: sudo bash setup.sh"
command -v dnf >/dev/null 2>&1 || die "dnf not found (this script targets Fedora/RHEL)"
command -v firewall-cmd >/dev/null 2>&1 || dnf install -y firewalld
systemctl enable --now firewalld >/dev/null

WAN_INTERFACE="$(ip -4 route show default | awk '{print $5; exit}')"
[ -n "$WAN_INTERFACE" ] || die "Could not detect the WAN interface (no default route)"

WAN_ZONE="$(firewall-cmd --get-zone-of-interface="$WAN_INTERFACE" 2>/dev/null || true)"
if [ -z "$WAN_ZONE" ] || [ "$WAN_ZONE" = "no zone" ]; then
    WAN_ZONE="$(firewall-cmd --get-default-zone)"
    warn "WAN interface has no explicit zone; using default zone '$WAN_ZONE'"
fi

echo
echo "VPN interface : $VPN_INTERFACE ($VPN_SERVER_IP)"
echo "WireGuard port: ${WG_PORT}/udp"
echo "WAN interface : $WAN_INTERFACE (zone: $WAN_ZONE)"
echo

# --- 1. WireGuard tools ------------------------------------------------------
if ! command -v wg >/dev/null 2>&1; then
    log "Installing wireguard-tools..."
    dnf install -y wireguard-tools
else
    log "wireguard-tools already installed"
fi

# --- 2. IPv4 forwarding ------------------------------------------------------
log "Enabling IPv4 forwarding..."
echo 'net.ipv4.ip_forward=1' > /etc/sysctl.d/99-vpn-manager.conf
sysctl -w net.ipv4.ip_forward=1 >/dev/null
sysctl --system >/dev/null

# --- 3. WireGuard interface config (only if missing) --------------------------
WG_CONF="/etc/wireguard/${VPN_INTERFACE}.conf"
if [ ! -f "$WG_CONF" ]; then
    log "Creating $WG_CONF (new server key pair)..."
    install -d -m 700 /etc/wireguard
    umask 077
    wg genkey | tee "/etc/wireguard/${VPN_INTERFACE}.key" | wg pubkey > "/etc/wireguard/${VPN_INTERFACE}.pub"
    cat > "$WG_CONF" <<EOF
[Interface]
Address = ${VPN_SERVER_IP}
ListenPort = ${WG_PORT}
PrivateKey = $(cat "/etc/wireguard/${VPN_INTERFACE}.key")
EOF
    chmod 600 "$WG_CONF"
else
    log "$WG_CONF already exists - leaving it untouched"
fi

systemctl enable "wg-quick@${VPN_INTERFACE}" >/dev/null 2>&1 || true
if ! ip link show "$VPN_INTERFACE" >/dev/null 2>&1; then
    log "Starting wg-quick@${VPN_INTERFACE}..."
    systemctl start "wg-quick@${VPN_INTERFACE}" || warn "Could not start wg-quick@${VPN_INTERFACE}; check: journalctl -u wg-quick@${VPN_INTERFACE}"
fi

# --- 4. firewalld: wireguard zone --------------------------------------------
if ! firewall-cmd --permanent --get-zones | tr ' ' '\n' | grep -qx wireguard; then
    log "Creating firewalld zone 'wireguard'..."
    firewall-cmd --permanent --new-zone=wireguard >/dev/null
fi

# Attach the VPN interface (permanent + runtime if it is up right now).
log "Attaching $VPN_INTERFACE to zone 'wireguard'..."
firewall-cmd --permanent --zone=wireguard --add-interface="$VPN_INTERFACE" >/dev/null 2>&1 || \
    firewall-cmd --permanent --zone=wireguard --change-interface="$VPN_INTERFACE" >/dev/null 2>&1 || \
    warn "Could not bind $VPN_INTERFACE to zone 'wireguard' (it may already be bound elsewhere)"

# SSH inside the tunnel (the manager app reads stats via the gateway IP).
log "Allowing SSH on zone 'wireguard'..."
firewall-cmd --permanent --zone=wireguard --add-service=ssh >/dev/null 2>&1 || true

# --- 5. WireGuard UDP port on WAN --------------------------------------------
log "Opening ${WG_PORT}/udp on zone '$WAN_ZONE'..."
firewall-cmd --permanent --zone="$WAN_ZONE" --add-port="${WG_PORT}/udp" >/dev/null 2>&1 || true

# --- 6. NAT + forwarding for full tunnel -------------------------------------
log "Enabling masquerade on zone '$WAN_ZONE'..."
firewall-cmd --permanent --zone="$WAN_ZONE" --add-masquerade >/dev/null 2>&1 || true

if firewall-cmd --permanent --get-policies >/dev/null 2>&1; then
    if ! firewall-cmd --permanent --get-policies | tr ' ' '\n' | grep -qx "$POLICY_NAME"; then
        log "Creating forwarding policy '$POLICY_NAME' (wireguard -> $WAN_ZONE)..."
        firewall-cmd --permanent --new-policy "$POLICY_NAME" >/dev/null
        firewall-cmd --permanent --policy "$POLICY_NAME" --add-ingress-zone=wireguard >/dev/null
        firewall-cmd --permanent --policy "$POLICY_NAME" --add-egress-zone="$WAN_ZONE" >/dev/null
        firewall-cmd --permanent --policy "$POLICY_NAME" --set-target=ACCEPT >/dev/null
    fi
else
    warn "This firewalld has no policy support; relying on masquerade only"
fi

# --- 7. SSH daemon -----------------------------------------------------------
systemctl enable --now sshd >/dev/null 2>&1 || true

# --- 8. Apply ----------------------------------------------------------------
log "Reloading firewalld..."
firewall-cmd --reload >/dev/null

echo
echo "==================== Done ===================="
echo "  - IPv4 forwarding   : $(sysctl -n net.ipv4.ip_forward)"
echo "  - $VPN_INTERFACE zone       : $(firewall-cmd --get-zone-of-interface="$VPN_INTERFACE" 2>/dev/null || echo 'interface not up yet')"
echo "  - UDP ${WG_PORT} open on  : $WAN_ZONE"
echo "  - masquerade on     : $WAN_ZONE ($WAN_INTERFACE)"
if [ -f "/etc/wireguard/${VPN_INTERFACE}.pub" ]; then
    echo "  - Server public key : $(cat "/etc/wireguard/${VPN_INTERFACE}.pub")"
fi
echo
echo "Verify:  sudo wg show $VPN_INTERFACE"
echo "         sudo firewall-cmd --zone=wireguard --list-all"
echo "Client:  curl ifconfig.me   (should print this server's public IP)"