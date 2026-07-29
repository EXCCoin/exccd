package peer

import (
	"net"
	"testing"
	"time"

	"github.com/EXCCoin/exccd/wire"
)

type addressedConn struct {
	net.Conn
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (c *addressedConn) LocalAddr() net.Addr  { return c.localAddr }
func (c *addressedConn) RemoteAddr() net.Addr { return c.remoteAddr }

func TestHandshakeDoneAfterVersionExchange(t *testing.T) {
	a, b := net.Pipe()
	addrA := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 19108}
	addrB := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 19109}
	inConn := &addressedConn{Conn: a, localAddr: addrA, remoteAddr: addrB}
	outConn := &addressedConn{Conn: b, localAddr: addrB, remoteAddr: addrA}

	cfg := &Config{Net: wire.TestNet3}
	inPeer := NewInboundPeer(cfg)
	outPeer, err := NewOutboundPeer(cfg, addrA.String())
	if err != nil {
		t.Fatal(err)
	}
	inPeer.AssociateConnection(inConn)
	outPeer.AssociateConnection(outConn)
	defer func() {
		inPeer.Disconnect()
		outPeer.Disconnect()
		inPeer.WaitForDisconnect()
		outPeer.WaitForDisconnect()
	}()

	deadline := time.After(time.Second)
	for !inPeer.HandshakeDone() || !outPeer.HandshakeDone() {
		select {
		case <-deadline:
			t.Fatal("version exchange did not complete")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
