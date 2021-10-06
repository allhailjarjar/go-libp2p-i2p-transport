package i2p

import (
	"net"
	"time"

	"github.com/joomcode/errorx"
	ma "github.com/multiformats/go-multiaddr"
)

type netAddr struct {
	str string
}

func (n *netAddr) String() string {
	return ""
}

func (n *netAddr) Network() string {
	return "I2P"
}

// ConnWithoutAddr is a net.Conn like but without LocalAddr and RemoteAddr.
type ConnWithoutAddr interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

//This struct only exists to satify libp2p interfaces
type Connection struct {
	ConnWithoutAddr

	localAddr  ma.Multiaddr
	remoteAddr ma.Multiaddr

	localNetAddr  net.Addr
	remoteNetAddr net.Addr
}

func NewConnection(conn ConnWithoutAddr, localAddr, remoteAddr ma.Multiaddr) (*Connection, error) {
	localNetAddrStr, err := MultiAddrToI2PAddr(localAddr) //manet.ToNetAddr(localAddr)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to convert MultiAddr to NetAddr")
	}

	remoteNetAddrStr, err := MultiAddrToI2PAddr(remoteAddr) //manet.ToNetAddr(remoteAddr)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to convert MultiAddr to NetAddr")
	}

	return &Connection{
		ConnWithoutAddr: conn,
		localAddr:       localAddr,
		remoteAddr:      remoteAddr,
		localNetAddr:    &netAddr{localNetAddrStr},
		remoteNetAddr:   &netAddr{remoteNetAddrStr},
	}, nil
}

//I don't think these are used anywhere.. but must match the interface
func (c *Connection) LocalAddr() net.Addr {
	return c.localNetAddr
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.remoteNetAddr
}

func (c *Connection) RemoteMultiaddr() ma.Multiaddr {
	return c.remoteAddr
}

func (c *Connection) LocalMultiaddr() ma.Multiaddr {
	return c.localAddr
}
