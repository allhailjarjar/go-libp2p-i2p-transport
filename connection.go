package i2p

import (
	"net"
	"time"

	"github.com/joomcode/errorx"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// ConnWithoutAddr is a net.Conn like but without LocalAddr and RemoteAddr.
type ConnWithoutAddr interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type ListenConnection struct {
}

type Connection struct {
	ConnWithoutAddr

	localAddr  ma.Multiaddr
	remoteAddr ma.Multiaddr

	localNetAddr  net.Addr
	remoteNetAddr net.Addr
	//todo: add listener
}

func NewConnection(conn ConnWithoutAddr, localAddr, remoteAddr ma.Multiaddr /*add listener here*/) (*Connection, error) {
	localNetAddr, err := manet.ToNetAddr(localAddr)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to convert MultiAddr to NetAddr")
	}

	remoteNetAddr, err := manet.ToNetAddr(remoteAddr)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to convert MultiAddr to NetAddr")
	}

	return &Connection{
		ConnWithoutAddr: conn,
		localAddr:       localAddr,
		remoteAddr:      remoteAddr,
		localNetAddr:    localNetAddr,
		remoteNetAddr:   remoteNetAddr,
	}, nil
}

//I don't think these are used anywhere, so we'll just stub them out
//They aren't very meaningful
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
