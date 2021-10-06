package i2p

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	csms "github.com/libp2p/go-conn-security-multistream"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	mplex "github.com/libp2p/go-libp2p-mplex"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const SAMHost = "127.0.0.1:7656"

/*********************************************

	Mocking out libp2p interfaces with a NOOP secure muxer

**********************************************/
type connSecurity struct {
	net.Conn
	privateKey      ic.PrivKey
	peerId          peer.ID
	remotePeer      peer.ID
	remotePublicKey ic.PubKey
}

func (c *connSecurity) LocalPeer() peer.ID {
	return c.peerId
}
func (c *connSecurity) LocalPrivateKey() ic.PrivKey {
	return c.privateKey
}
func (c *connSecurity) RemotePeer() peer.ID {
	return c.remotePeer
}
func (c *connSecurity) RemotePublicKey() ic.PubKey {
	return c.remotePublicKey
}

type secureMuxer struct {
	remotePeerId    peer.ID
	remotePublicKey ic.PubKey
	privateKey      ic.PrivKey
}

func (s *secureMuxer) SecureInbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, bool, error) {
	sConn := &connSecurity{
		insecure,
		s.privateKey,
		p,
		s.remotePeerId,
		s.remotePublicKey,
	}

	return sConn, true, nil
}

func (s *secureMuxer) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, bool, error) {
	sConn := &connSecurity{
		insecure,
		s.privateKey,
		p,
		s.remotePeerId,
		s.remotePublicKey,
	}

	return sConn, false, nil
}

func makeInsecureMuxerMocked(t *testing.T) (peer.ID, sec.SecureMuxer) {
	t.Helper()
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	require.NoError(t, err)
	mockedRmotePriv, mockedRmotePub, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	mockedRemoteID, err := peer.IDFromPrivateKey(mockedRmotePriv)
	require.NoError(t, err)

	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	secMuxer := &secureMuxer{
		mockedRemoteID,
		mockedRmotePub,
		priv,
	}
	return id, secMuxer
}

/************************************
	END mocking libp2p interfaces

*************************************/

func makeInsecureMuxer(t *testing.T) (peer.ID, sec.SecureMuxer) {
	t.Helper()
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	require.NoError(t, err)

	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	var secMuxer csms.SSMuxer
	secMuxer.AddTransport(insecure.ID, insecure.NewWithIdentity(id, priv))
	return id, &secMuxer
}

func TestBuildI2PTransport(t *testing.T) {
	ch := make(chan string, 1)
	go setupServer(t, ch)

	serverAddr := <-ch
	setupClient(t, i2pkeys.I2PAddr(serverAddr), 2345)

}

func setupClient(t *testing.T, serverAddr i2pkeys.I2PAddr, randNum int) {
	log.Println("Starting client setup")
	sam, err := sam3.NewSAM(SAMHost)
	if err != nil {
		assert.Fail(t, "Failed to connect to SAM")
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		assert.Fail(t, "Failed to generate keys")
		return
	}

	builder, _, err := I2PTransportBuilder(sam, keys, "23459", int(time.Now().UnixNano()))
	assert.NoError(t, err)

	peerID, sm := makeInsecureMuxerMocked(t)
	secureTransport, err := builder(&tptu.Upgrader{
		Secure: sm,
		Muxer:  new(mplex.Transport),
	})

	serverMultiAddr, err := I2PAddrToMultiAddr(string(serverAddr))
	log.Println("Dialing host on this destination: " + serverMultiAddr.String())

	for i := 0; i < 5; i++ {
		log.Println("Starting dial")
		conn, err := secureTransport.Dial(context.TODO(), serverMultiAddr, peerID)
		if err != nil {
			assert.Fail(t, "Failed to dial", err)
			return
		}
		log.Println("Opening Stream")
		stream, err := conn.OpenStream(context.TODO())
		if err != nil {
			assert.Fail(t, "Failed to open outbound stream", err)
			return
		}

		//stream.Write([]byte("Hello!"))

		// buf := make([]byte, 1024)
		// nBytes, err := stream.Read(buf)
		// if err != nil {
		// 	panic(err)
		// }
		//log.Println("Server output: " + string(buf[:nBytes]))
		stream.Close()
	}
}

func setupServer(t *testing.T, addrChan chan string) {
	sam, err := sam3.NewSAM(SAMHost)
	if err != nil {
		assert.Fail(t, "Failed to connect to SAM", err)
		addrChan <- ""
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		assert.Fail(t, "Failed to generate keys", err)
		addrChan <- ""
		return
	}

	port := "45793"
	builder, listenAddr, err := I2PTransportBuilder(sam, keys, port, int(time.Now().UnixNano()))
	assert.NoError(t, err)

	_, sm := makeInsecureMuxerMocked(t)
	secureTransport, err := builder(&tptu.Upgrader{
		Secure: sm,
		Muxer:  new(mplex.Transport),
	})

	listener, err := secureTransport.Listen(listenAddr)
	if err != nil {
		assert.Fail(t, "Failed to create listener", err)
		addrChan <- ""
		return
	}

	addrChan <- listener.Addr().String()
	log.Println("Listener Addr: " + listener.Addr().String())

	for i := 0; i < 5; i++ {
		capableConnection, err := listener.Accept()
		if err != nil {
			assert.Fail(t, "Failed to accept connection: "+err.Error())
		}

		stream, err := capableConnection.AcceptStream()

		// buf := make([]byte, 1024)
		// _, err = stream.Read(buf)
		// stream.Write([]byte(capableConnection.LocalMultiaddr().String()))
		//log.Println(capableConnection.RemoteMultiaddr())

		stream.Close()
	}

}
