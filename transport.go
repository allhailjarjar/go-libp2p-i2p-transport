package i2p

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/eyedeekay/sam3/i2pkeys"
	"github.com/joomcode/errorx"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"

	"github.com/eyedeekay/sam3"
)

type I2PTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader *tptu.Upgrader

	sam             *sam3.SAM
	i2PKeys         i2pkeys.I2PKeys
	primarySession  *sam3.PrimarySession
	outboundSession *sam3.StreamSession
	inboundSession  *sam3.StreamSession
	//sync.RWMutex
}

var _ transport.Transport = &I2PTransport{}

type Option func(*I2PTransport) error

type TransportBuilderFunc = func(*tptu.Upgrader) (*I2PTransport, error)

//returns a function that when called by go-libp2p, creates an I2PTransport
//Initializes SAM sessions/tunnel which can take about 4-25 seconds depending
//on i2p network conditions
func I2PTransportBuilder(sam *sam3.SAM,
	i2pKeys i2pkeys.I2PKeys, outboundPort string, rngSeed int) (TransportBuilderFunc, ma.Multiaddr, error) {
	rand.Seed(int64(rngSeed))

	randSessionSuffix := strconv.Itoa(rand.Int())

	samPrimarySession, err := sam.NewPrimarySession("primarySession-"+randSessionSuffix, i2pKeys, sam3.Options_Default)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create Primary session with I2P SAM")
	}

	inboundSession, err := samPrimarySession.NewStreamSubSession("inboundSession-" + randSessionSuffix)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create inboundSession subsession with I2P SAM")
	}

	outboundSession, err := samPrimarySession.NewStreamSubSessionWithPorts("outboundSession-"+randSessionSuffix, outboundPort, "0")
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create outbound subsession with I2P SAM")
	}

	i2pDestination, err := I2PAddrToMultiAddr(string(samPrimarySession.Addr()))
	if err != nil {
		return nil, nil, err
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
		return i2p, nil

	}, i2pDestination, nil
}

var dialMatcher = mafmt.Or(
	mafmt.Base(ma.P_GARLIC64),
	mafmt.Base(ma.P_GARLIC32),
)

// CanDial returns true if this transport believes it can dial the given multiaddr.
func (i2p *I2PTransport) CanDial(addr ma.Multiaddr) bool {
	return dialMatcher.Matches(addr)
}

func (i2p *I2PTransport) Dial(ctx context.Context, remoteAddress ma.Multiaddr, peerID peer.ID) (transport.CapableConn, error) {
	//In case libp2p tries to dial a non-garlic address, we should error early
	if !i2p.CanDial(remoteAddress) {
		return nil, errorx.IllegalArgument.New(fmt.Sprintf("Can't dial \"%s\"", remoteAddress))
	}

	remoteNetAddr, err := MultiAddrToI2PAddr(remoteAddress)
	if err != nil {
		return nil, err
	}

	conn, err := i2p.outboundSession.DialI2P(i2pkeys.I2PAddr(remoteNetAddr))
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to dial remote address")
	}

	localAddress, err := I2PAddrToMultiAddr(i2p.outboundSession.LocalAddr().String()) //manet.FromNetAddr(i2p.outboundSession.LocalAddr())
	if err != nil {
		return nil, errorx.Decorate(err, "Unable to constuct multi-addr from net address")
	}

	outboundConnection, err := NewConnection(conn, localAddress, remoteAddress)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to construct Connection type")
	}

	return i2p.Upgrader.Upgrade(ctx, i2p, outboundConnection, network.DirOutbound, peerID)
	//return i2p.Upgrader.UpgradeOutbound(ctx, i2p, outboundConnection, peerID)
}

//input argument isn't used because we'll be listening on whichever destination is provided
//by i2p
func (i2p *I2PTransport) Listen(_ ma.Multiaddr) (transport.Listener, error) {
	streamListener, err := i2p.outboundSession.Listen()
	if err != nil {
		return nil, errorx.Decorate(err, "Unable to call listen on SAM session")
	}

	listener, err := NewTransportListener(streamListener)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to nitialize transport listener")
	}

	return i2p.Upgrader.UpgradeListener(i2p, listener), nil
}

//Closes all SAM sessions by closing the PRIMARY session
func (i2p *I2PTransport) Close() {
	i2p.primarySession.Close()
}

// Protocols returns the list of protocols this transport can dial.
func (i2p *I2PTransport) Protocols() []int {
	return []int{ma.P_GARLIC64, ma.P_GARLIC32, ma.P_TCP}
}

// Proxy always returns false for the I2P transport.
func (i2p *I2PTransport) Proxy() bool {
	return false
}

func (i2p *I2PTransport) String() string {
	return "I2P"
}
