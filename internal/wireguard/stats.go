package wireguard

import (
	"strconv"
	"strings"
	"time"

	"vpn-manager/internal/data"
	sshclient "vpn-manager/internal/ssh"
)

// ServerStats holds aggregated peer statistics from the server.
type ServerStats struct {
	PeerCount     int
	TotalRx       uint64
	TotalTx       uint64
	LastHandshake time.Time
}

// GetServerStats queries the server for aggregate peer statistics.
func GetServerStats() (ServerStats, error) {
	var stats ServerStats

	settings, err := data.ReadSettingsFile()
	if err != nil {
		return stats, err
	}

	if err := validateInterface(settings.ServerInterface); err != nil {
		return stats, err
	}

	// Only aggregate traffic for peers that belong to this app, so the
	// dashboard does not show other tunnels (for example a manually added peer).
	devices, err := data.ReadDevicesFile()
	if err != nil {
		devices = []data.Device{}
	}
	ourKeys := make(map[string]bool, len(devices))
	for _, d := range devices {
		ourKeys[d.PublicKey] = true
	}

	client, err := sshclient.ConnectServer()
	if err != nil {
		return stats, err
	}
	defer client.Close()

	output, err := sshclient.RunCommand(client, "sudo wg show "+settings.ServerInterface+" dump")
	if err != nil {
		return stats, err
	}

	for i, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if i == 0 {
			// The first line describes the interface, not a peer
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}

		// Skip peers that are not managed by this app
		if !ourKeys[fields[0]] {
			continue
		}

		rx, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			continue
		}
		tx, err := strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			continue
		}

		stats.PeerCount++
		stats.TotalRx += rx
		stats.TotalTx += tx

		if hs, err := strconv.ParseInt(fields[4], 10, 64); err == nil && hs > 0 {
			t := time.Unix(hs, 0)
			if t.After(stats.LastHandshake) {
				stats.LastHandshake = t
			}
		}
	}

	return stats, nil
}
