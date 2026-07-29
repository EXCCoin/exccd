package peer

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/EXCCoin/exccd/wire"
)

type deadlineConn struct {
	net.Conn
	writeDeadline time.Time
}

func (c *deadlineConn) SetWriteDeadline(deadline time.Time) error {
	c.writeDeadline = deadline
	return nil
}

type maxSizedMsg struct {
	maxPayload uint32
}

func (*maxSizedMsg) BtcDecode(io.Reader, uint32) error { return nil }
func (*maxSizedMsg) BtcEncode(io.Writer, uint32) error { return nil }
func (*maxSizedMsg) Command() string                   { return "sized" }
func (msg *maxSizedMsg) MaxPayloadLength(uint32) uint32 {
	return msg.maxPayload
}

type serializedMsg struct {
	maxSizedMsg
	serializedSize int
}

func (msg *serializedMsg) SerializeSize() int { return msg.serializedSize }

func TestWriteMessageSetsDeadline(t *testing.T) {
	local, remote := net.Pipe()
	defer local.Close()
	defer remote.Close()
	go io.Copy(io.Discard, remote)

	conn := &deadlineConn{Conn: local}
	p := NewInboundPeer(&Config{})
	p.conn = conn
	tests := []struct {
		name          string
		msg           wire.Message
		wantAllowance time.Duration
	}{
		{
			name: "maximum payload fallback",
			msg: &maxSizedMsg{
				maxPayload: 2*writeStallBytesPerSec - wire.MessageHeaderSize,
			},
			wantAllowance: 2 * time.Second,
		},
		{
			name: "serialized payload size",
			msg: &serializedMsg{
				maxSizedMsg: maxSizedMsg{
					maxPayload: 4*writeStallBytesPerSec - wire.MessageHeaderSize,
				},
				serializedSize: 2*writeStallBytesPerSec - wire.MessageHeaderSize,
			},
			wantAllowance: 2 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			earliest := time.Now().Add(writeStallTimeout + test.wantAllowance)
			if err := p.writeMessage(test.msg); err != nil {
				t.Fatal(err)
			}
			latest := time.Now().Add(writeStallTimeout + test.wantAllowance)
			if conn.writeDeadline.Before(earliest) ||
				conn.writeDeadline.After(latest) {

				t.Fatalf("write deadline %v outside [%v, %v]",
					conn.writeDeadline, earliest, latest)
			}
		})
	}
}
