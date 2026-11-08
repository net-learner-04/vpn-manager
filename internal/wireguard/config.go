package wireguard

import (
	"bytes"
	"net"
	"text/template"

	"vpn-manager/internal/data"
)

type ClientConfig struct {
	PrivateKey      string
	Address         string
	DNS             string
	ServerPublicKey string
	Endpoint        string
	AllowedIPs      string
}

var ConfigTemplate = template.Must(template.New("client").Parse(`[Interface]
PrivateKey = {{.PrivateKey}}
Address = {{.Address}}
DNS = {{.DNS}}

[Peer]
PublicKey = {{.ServerPublicKey}}
Endpoint = {{.Endpoint}}
AllowedIPs = {{.AllowedIPs}}
PersistentKeepalive = 25
`))

func RenderClientConfig(c ClientConfig) ([]byte, error) {
	var buf bytes.Buffer
	if err := ConfigTemplate.Execute(&buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Build a ClientConfig from the saved settings and a device.
func BuildClientConfig(s data.Settings, d data.Device, privateKey string) ClientConfig {
	return ClientConfig{
		PrivateKey:      privateKey,
		Address:         d.IP + "/32",
		DNS:             s.DNS,
		ServerPublicKey: s.ServerPublicKey,
		Endpoint:        net.JoinHostPort(s.ServerAddress, s.WireGuardPort),
		AllowedIPs:      s.AllowedIPs,
	}
}
