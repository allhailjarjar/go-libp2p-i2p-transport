package i2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const base64Addr = "jT~IyXaoauTni6N4517EG8mrFUKpy0IlgZh-EY9csMAk82Odatmzr~YTZy8Hv7u~wvkg75EFNOyqb~nAPg-khyp2TS~ObUz8WlqYAM2VlEzJ7wJB91P-cUlKF18zSzVoJFmsrcQHZCirSbWoOknS6iNmsGRh5KVZsBEfp1Dg3gwTipTRIx7Vl5Vy~1OSKQVjYiGZS9q8RL0MF~7xFiKxZDLbPxk0AK9TzGGqm~wMTI2HS0Gm4Ycy8LYPVmLvGonIBYndg2bJC7WLuF6tVjVquiokSVDKFwq70BCUU5AU-EvdOD5KEOAM7mPfw-gJUG4tm1TtvcobrObqoRnmhXPTBTN5H7qDD12AvlwFGnfAlBXjuP4xOUAISL5SRLiulrsMSiT4GcugSI80mF6sdB0zWRgL1yyvoVWeTBn1TqjO27alr95DGTluuSqrNAxgpQzCKEWAyzrQkBfo2avGAmmz2NaHaAvYbOg0QSJz1PLjv2jdPW~ofiQmrGWM1cd~1cCqAAAA"
const base32AddrSuffix = "ugbgtbk6qvbymwgv2clzeefcxrjz4milklcyi6hzqxmcxxnwjh5a.b32.i2p"
const base32Addr = "ugbgtbk6qvbymwgv2clzeefcxrjz4milklcyi6hzqxmcxxnwjh5a"

func TestNetAddrToGarlic64MultiAddr(t *testing.T) {
	multiAddr, err := I2PAddrToMultiAddr(base32AddrSuffix)
	assert.NoError(t, err)
	assert.Equal(t, "/garlic32/"+base32Addr, multiAddr.String())

	multiAddr2, err := I2PAddrToMultiAddr(base32Addr)
	assert.NoError(t, err)
	assert.Equal(t, "/garlic32/"+base32Addr, multiAddr2.String())

	multiAddr3, err := I2PAddrToMultiAddr(base64Addr)
	assert.NoError(t, err)
	assert.Equal(t, "/garlic64/"+base64Addr, multiAddr3.String())

}

func TestMultiAddrToI2PAddr(t *testing.T) {
	multiAddrB32, err := I2PAddrToMultiAddr(base32Addr)
	assert.NoError(t, err)

	addr, err := MultiAddrToI2PAddr(multiAddrB32)
	assert.NoError(t, err)
	assert.Equal(t, base32AddrSuffix, addr)

	multiAddrB64, err := I2PAddrToMultiAddr(base64Addr)
	assert.NoError(t, err)

	addr2, err := MultiAddrToI2PAddr(multiAddrB64)
	assert.NoError(t, err)
	assert.Equal(t, base64Addr, addr2)

}
