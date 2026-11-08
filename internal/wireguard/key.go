package wireguard

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Keys struct {
	Private string
	Public  string
}

func GenerateKeyPair() (Keys, error) {
	k, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return Keys{}, err
	}

	return Keys{
		Private: k.String(),
		Public:  k.PublicKey().String(),
	}, nil
}
