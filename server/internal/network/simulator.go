package network

import (
	"context"
	"net"
	"time"

	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
)

// SimulatedSender wraps an existing Server to inject artificial latency into outgoing packets.
type SimulatedSender struct {
	inner Server
	delay time.Duration
}

func NewSimulatedSender(inner Server, delayMs int) *SimulatedSender {
	return &SimulatedSender{
		inner: inner,
		delay: time.Duration(delayMs) * time.Millisecond,
	}
}

func (s *SimulatedSender) Send(target net.Addr, pkt packet.Packet) error {
	if s.delay <= 0 {
		return s.inner.Send(target, pkt)
	}

	// Capture values for the goroutine
	t := target
	p := pkt
	go func() {
		time.Sleep(s.delay)
		_ = s.inner.Send(t, p)
	}()

	return nil
}

// SimulatedHandler wraps an existing PacketHandler to inject artificial latency into incoming packets.
type SimulatedHandler struct {
	inner PacketHandler
	delay time.Duration
}

func NewSimulatedHandler(inner PacketHandler, delayMs int) *SimulatedHandler {
	return &SimulatedHandler{
		inner: inner,
		delay: time.Duration(delayMs) * time.Millisecond,
	}
}

func (h *SimulatedHandler) HandlePacket(ctx context.Context, source net.Addr, pkt packet.Packet) {
	if h.delay <= 0 {
		h.inner.HandlePacket(ctx, source, pkt)
		return
	}

	go func() {
		time.Sleep(h.delay)
		h.inner.HandlePacket(ctx, source, pkt)
	}()
}
