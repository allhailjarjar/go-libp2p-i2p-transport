package i2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/eyedeekay/sam3/i2pkeys"
	logging "github.com/ipfs/go-log/v2"
	"github.com/joomcode/errorx"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/eyedeekay/sam3"
)

var log = logging.Logger("i2p-tpt")

type I2PTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader *tptu.Upgrader

	sam             *sam3.SAM
	i2PKeys         i2pkeys.I2PKeys
	primarySession  *sam3.PrimarySession
	outboundSession *sam3.StreamSession
	inboundSession  *sam3.StreamSession
	sync.RWMutex
}

var _ transport.Transport = &I2PTransport{}

type Option func(*I2PTransport) error

type TransportBuilderFunc = func(*tptu.Upgrader) (*I2PTransport, error)

//returns a function that when called by go-libp2p, creates an I2PTransport
func I2PTransportBuilder(sam *sam3.SAM,
	i2pKeys i2pkeys.I2PKeys, opts ...Option) (TransportBuilderFunc, error) {

	samPrimarySession, err := sam.NewPrimarySession("primarySession", i2pKeys, sam3.Options_Default)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to create Primary session with I2P SAM")
	}

	outboundSession, err := samPrimarySession.NewStreamSubSession("outboundSession")
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to create outbound subsession with I2P SAM")
	}

	inboundSession, err := samPrimarySession.NewStreamSubSession("inboundSession")
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to create inboundSession subsession with I2P SAM")
	}

	return func(upgrader *tptu.Upgrader) (*I2PTransport, error) {
		i2p := &I2PTransport{
			Upgrader:        upgrader,
			sam:             sam,
			i2PKeys:         i2pKeys,
			primarySession:  samPrimarySession,
			outboundSession: outboundSession,
			inboundSession:  inboundSession,
		}

		for _, o := range opts {
			if err := o(i2p); err != nil {
				return nil, err
			}
		}
		return i2p, nil

	}, nil
}

var dialMatcher = mafmt.Base(ma.P_GARLIC64)

// CanDial returns true if this transport believes it can dial the given multiaddr.
func (i2p *I2PTransport) CanDial(addr ma.Multiaddr) bool {
	return dialMatcher.Matches(addr)
}

func (i2p *I2PTransport) Dial(ctx context.Context, remoteAddress ma.Multiaddr, peerID peer.ID) (transport.CapableConn, error) {
	//In case libp2p tries to dial a non-garlic address, we should error early
	if !i2p.CanDial(remoteAddress) {
		return nil, errorx.IllegalArgument.New(fmt.Sprintf("Can't dial \"%s\"", remoteAddress))
	}

	remoteNetAddr, err := multiAddrToNetAddr(remoteAddress)
	if err != nil {
		return nil, err
	}

	conn, err := i2p.outboundSession.DialContext(ctx, "tcp", remoteNetAddr)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to dial remote address")
	}

	localAddress, err := manet.FromNetAddr(i2p.outboundSession.LocalAddr())
	if err != nil {
		return nil, errorx.Decorate(err, "Unable to constuct multi-addr from net address")
	}

	outboundConnection, err := NewConnection(conn, localAddress, remoteAddress)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to construct Connection type")
	}

	return i2p.Upgrader.UpgradeOutbound(ctx, i2p, outboundConnection, peerID)
}

func (i2p *I2PTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	return nil, nil
}

//Closes all SAM sessions by closing the PRIMARY session
func (i2p *I2PTransport) Close() {
	i2p.primarySession.Close()
}

// Protocols returns the list of terminal protocols this transport can dial.
func (i2p *I2PTransport) Protocols() []int {
	return []int{ma.P_GARLIC64}
}

// Proxy always returns false for the I2P transport.
func (i2p *I2PTransport) Proxy() bool {
	return false
}

func (i2p *I2PTransport) String() string {
	return "I2P"
}
