package i2p

import (
	"errors"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

func MultiAddrToI2PAddr(addr ma.Multiaddr) (string, error) {
	numProtocols := len(addr.Protocols())
	if numProtocols != 1 {
		return "", errors.New(fmt.Sprintf("Expected 1 protocols in multiaddr but found %d", numProtocols))
	}

	destination, err := addr.ValueForProtocol(addr.Protocols()[0].Code)
	if err != nil {
		return "", err
	}

	if len(destination) <= 55 {
		destination += ".b32.i2p"
	}

	return destination, nil
}

//expects either a base32 or base64 i2p destination
//expects there to be no :port suffix to the address
func I2PAddrToMultiAddr(addr string) (ma.Multiaddr, error) {
	if len(addr) < 52 {
		return nil, errors.New("Address too short for a i2p")
	}

	garlicBase := "/garlic64/"

	//handle base32 destinations
	//55 for max address and 8 extra for .b32.i2p suffix
	if len(addr) <= 63 {
		//check to see if the address has a .b32.i2p suffix
		//if exists, remove
		if addr[len(addr)-8:] == ".b32.i2p" {
			addr = addr[:len(addr)-8]
		}
		garlicBase = "/garlic32/"
	}

	return ma.NewMultiaddr(garlicBase + addr)
}
