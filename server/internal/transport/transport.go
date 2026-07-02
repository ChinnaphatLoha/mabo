package transport

import (
	"context"
	"net"

	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
)

type ClientID string

type IncomingPacket struct {
	Source net.Addr
	Packet packet.Packet
	RecvAt int64
}

type OutgoingPacket struct {
	Target net.Addr
	Packet packet.Packet
	SendAt int64
}

type Transport interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, target net.Addr, pkt packet.Packet) error
}

func NewClientID(addr net.Addr) ClientID {
	return ClientID(addr.String())
}
