package i2p

import (
	"errors"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

const I2PProtocol = ma.P_GARLIC64

func multiAddrToNetAddr(addr ma.Multiaddr) (string, error) {
	numProtocols := len(addr.Protocols())
	if numProtocols < 1 {
		return "", errors.New(fmt.Sprintf("Expected 1 protocols in multiaddr but found %d", numProtocols))
	}

	return addr.ValueForProtocol(ma.P_GARLIC64)
}
