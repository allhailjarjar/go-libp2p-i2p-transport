package i2p

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	csms "github.com/libp2p/go-conn-security-multistream"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	mplex "github.com/libp2p/go-libp2p-mplex"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const SAMHost = "127.0.0.1:7656"

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

type ServerInfo struct {
	Addr   i2pkeys.I2PAddr
	PeerID peer.ID
}

func TestBuildI2PTransport(t *testing.T) {
	ch := make(chan *ServerInfo, 1)
	go setupServer(t, ch)

	serverAddrAndPeer := <-ch
	setupClient(t, serverAddrAndPeer.Addr, serverAddrAndPeer.PeerID, 2345)

}

func setupClient(t *testing.T, serverAddr i2pkeys.I2PAddr, serverPeerID peer.ID, randNum int) {
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

	peerID, sm := makeInsecureMuxer(t)
	log.Println("Client Peer ID is: " + peerID.String())
	secureTransport, err := builder(&tptu.Upgrader{
		Secure: sm,
		Muxer:  new(mplex.Transport),
	})

	serverMultiAddr, err := I2PAddrToMultiAddr(string(serverAddr))
	log.Println("Dialing host on this destination: " + serverMultiAddr.String())

	for i := 0; i < 5; i++ {
		log.Println("Starting dial")
		conn, err := secureTransport.Dial(context.TODO(), serverMultiAddr, serverPeerID)
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

		stream.Write([]byte("Hello!"))
		stream.Close()
	}
}

func setupServer(t *testing.T, addrChan chan *ServerInfo) {
	sam, err := sam3.NewSAM(SAMHost)
	if err != nil {
		assert.Fail(t, "Failed to connect to SAM", err)
		addrChan <- nil
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		assert.Fail(t, "Failed to generate keys", err)
		addrChan <- nil
		return
	}

	port := "45793"
	builder, listenAddr, err := I2PTransportBuilder(sam, keys, port, int(time.Now().UnixNano()))
	assert.NoError(t, err)

	peerID, sm := makeInsecureMuxer(t)
	log.Println("Server Peer ID is: " + peerID.String())

	secureTransport, err := builder(&tptu.Upgrader{
		Secure: sm,
		Muxer:  new(mplex.Transport),
	})

	listener, err := secureTransport.Listen(listenAddr)
	if err != nil {
		assert.Fail(t, "Failed to create listener", err)
		addrChan <- nil
		return
	}

	serverInfo := &ServerInfo{
		i2pkeys.I2PAddr(listener.Addr().String()),
		peerID,
	}
	addrChan <- serverInfo
	log.Println("Listener Addr: " + listener.Addr().String())

	for i := 0; i < 5; i++ {
		capableConnection, err := listener.Accept()
		if err != nil {
			assert.Fail(t, "Failed to accept connection: "+err.Error())
		}

		stream, err := capableConnection.AcceptStream()

		buf := make([]byte, 1024)
		_, err = stream.Read(buf)
		stream.Write([]byte(capableConnection.LocalMultiaddr().String()))
		log.Println(capableConnection.RemoteMultiaddr())

		stream.Close()
	}

}
