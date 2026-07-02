package network

import (
	"context"
	"net"

	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
)

type DefaultHandler struct {
	logger *logger.Logger
}

func NewDefaultHandler(logger *logger.Logger) *DefaultHandler {
	return &DefaultHandler{logger: logger}
}

func (h *DefaultHandler) HandlePacket(ctx context.Context, source net.Addr, pkt packet.Packet) {
	h.logger.Debug("received packet", "source", source.String(), "packet_id", pkt.ID.String())
}
