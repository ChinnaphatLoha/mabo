package network

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
)

type PacketHandler interface {
	HandlePacket(ctx context.Context, source net.Addr, pkt packet.Packet)
}

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(target net.Addr, pkt packet.Packet) error
}

type Config struct {
	BindAddress    string
	ReadBufferSize int
}

type UDPServer struct {
	cfg     Config
	handler PacketHandler
	logger  *logger.Logger
	conn    *net.UDPConn
}

func NewUDPServer(cfg Config, handler PacketHandler, logger *logger.Logger) *UDPServer {
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = 4096
	}
	return &UDPServer{
		cfg:     cfg,
		handler: handler,
		logger:  logger,
	}
}

func (s *UDPServer) Start(ctx context.Context) error {
	if s.handler == nil {
		return errors.New("packet handler is required")
	}

	addr, err := net.ResolveUDPAddr("udp", s.cfg.BindAddress)
	if err != nil {
		return err
	}

	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("network listener started",
		"address", s.cfg.BindAddress,
	)

	buffer := make([]byte, s.cfg.ReadBufferSize)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("network listener stopping")
			return nil
		default:
		}

		if err := s.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return err
		}

		n, remote, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return err
		}

		s.logger.Debug("udp packet received",
			"remote", remote.String(),
			"size", n,
		)

		if n == 0 {
			continue
		}

		pkt := packet.Packet{
			ID:      packet.ID(buffer[0]),
			Payload: append([]byte(nil), buffer[1:n]...),
		}

		go s.handler.HandlePacket(ctx, remote, pkt)
	}
}

func (s *UDPServer) Stop(ctx context.Context) error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func (s *UDPServer) Send(target net.Addr, pkt packet.Packet) error {
	if s.conn == nil {
		return errors.New("network not started")
	}

	// Build buffer: first byte is packet ID, remaining is payload
	buf := make([]byte, 1+len(pkt.Payload))
	buf[0] = byte(pkt.ID)
	copy(buf[1:], pkt.Payload)

	udpAddr, ok := target.(*net.UDPAddr)
	if !ok {
		// Try to resolve from string
		var err error
		udpAddr, err = net.ResolveUDPAddr("udp", target.String())
		if err != nil {
			return err
		}
	}

	_, err := s.conn.WriteToUDP(buf, udpAddr)
	return err
}
