package app

import (
	"context"
	"errors"
	"net"

	"github.com/ChinnaphatLoha/mabo/server/internal/command"
	"github.com/ChinnaphatLoha/mabo/server/internal/game"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/match"
	"github.com/ChinnaphatLoha/mabo/server/internal/network"
	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
	"github.com/ChinnaphatLoha/mabo/server/internal/protocol"
	"github.com/ChinnaphatLoha/mabo/server/internal/room"
	"github.com/ChinnaphatLoha/mabo/server/internal/session"
	"github.com/ChinnaphatLoha/mabo/server/internal/world"
)

// Sender is a narrow interface so the handler never needs the full network.Server.
type Sender interface {
	Send(target net.Addr, pkt packet.Packet) error
}

// MultiplayerHandler implements network.PacketHandler.
// It owns application-level logic: decode → validate → call managers → respond.
type MultiplayerHandler struct {
	sessions *session.Manager
	rooms    *room.Manager
	matches  *match.Manager
	world    *world.World
	inputs   *command.InputBuffer
	sender   Sender
	logger   *logger.Logger
}

var _ network.PacketHandler = (*MultiplayerHandler)(nil)

// NewMultiplayerHandler creates a handler with all required managers wired.
func NewMultiplayerHandler(
	sessions *session.Manager,
	rooms *room.Manager,
	matches *match.Manager,
	w *world.World,
	inputs *command.InputBuffer,
	sender Sender,
	log *logger.Logger,
) *MultiplayerHandler {
	return &MultiplayerHandler{
		sessions: sessions,
		rooms:    rooms,
		matches:  matches,
		world:    w,
		inputs:   inputs,
		sender:   sender,
		logger:   log,
	}
}

// HandlePacket dispatches an incoming packet to the appropriate handler.
func (h *MultiplayerHandler) HandlePacket(ctx context.Context, source net.Addr, pkt packet.Packet) {
	switch pkt.ID {
	case packet.PacketIDConnect:
		// No-op; UDP has no real handshake beyond LoginRequest.
	case packet.PacketIDDisconnect:
		h.handleDisconnect(source)
	case packet.PacketIDLoginRequest:
		h.handleLogin(source, pkt)
	case packet.PacketIDCreateRoomRequest:
		h.handleCreateRoom(source, pkt)
	case packet.PacketIDJoinRoomRequest:
		h.handleJoinRoom(source, pkt)
	case packet.PacketIDLeaveRoomRequest:
		h.handleLeaveRoom(source, pkt)
	case packet.PacketIDMovementInput:
		h.handleMovementInput(source, pkt)
	default:
		h.logger.Debug("unknown packet id", "id", pkt.ID.String(), "source", source.String())
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidPacket, "unknown packet id")
	}
}

// ── Login ──────────────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleLogin(source net.Addr, pkt packet.Packet) {
	var req protocol.LoginRequest
	if err := protocol.Unmarshal(pkt.Payload, &req); err != nil {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidPacket, "bad json")
		return
	}

	s, err := h.sessions.Create(source, req.GuestName)
	if errors.Is(err, session.ErrAlreadyConnected) {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorAlreadyConnected, "already connected")
		return
	}
	if err != nil {
		h.logger.Error("create session", "error", err)
		return
	}

	h.send(source, packet.PacketIDLoginResponse, protocol.LoginResponse{
		SessionID: s.ID,
		PlayerID:  s.PlayerID,
	})
}

// ── Disconnect ─────────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleDisconnect(source net.Addr) {
	s, ok := h.sessions.RemoveByAddr(source)
	if !ok {
		return
	}
	h.cleanupPlayer(s.PlayerID)
}

// cleanupPlayer removes a player from room, match, and world, broadcasting
// PlayerDisconnected to any remaining room members.
func (h *MultiplayerHandler) cleanupPlayer(playerID string) {
	// Find room membership via a linear scan (acceptable for Day 2 scale).
	// In production a playerID→roomID reverse index would replace this.
	var targetRoom *room.Room
	for _, candidate := range h.roomsSnapshot() {
		players, _, _, _ := candidate.Snapshot()
		for _, p := range players {
			if p == playerID {
				targetRoom = candidate
				break
			}
		}
		if targetRoom != nil {
			break
		}
	}

	roomID := ""
	matchID := ""

	if targetRoom != nil {
		players, _, _, _ := targetRoom.Snapshot()
		roomID = targetRoom.ID

		// Broadcast PlayerDisconnected before we mutate membership.
		mt, hasMt := h.matches.GetByRoomID(roomID)
		if hasMt {
			matchID = mt.ID
		}

		msg := protocol.PlayerDisconnected{
			RoomID:   roomID,
			MatchID:  matchID,
			PlayerID: playerID,
		}
		for _, pid := range players {
			if pid == playerID {
				continue
			}
			if s, ok := h.sessions.GetByPlayerID(pid); ok {
				h.send(s.Addr, packet.PacketIDPlayerDisconnected, msg)
			}
		}

		targetRoom.Leave(playerID)
		if hasMt {
			mt.RemovePlayer(playerID)
		}
	}

	// Remove from world.
	h.world.RemovePlayer(playerID)
}

// roomsSnapshot returns all known rooms. Used by cleanupPlayer.
func (h *MultiplayerHandler) roomsSnapshot() []*room.Room {
	// room.Manager does not expose a List yet; we add a wrapper here.
	return h.rooms.List()
}

// ── Create Room ────────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleCreateRoom(source net.Addr, pkt packet.Packet) {
	s, ok := h.requireSession(source, pkt.ID)
	if !ok {
		return
	}

	var req protocol.CreateRoomRequest
	_ = protocol.Unmarshal(pkt.Payload, &req) // capacity omitempty – ok to ignore error

	capacity := req.Capacity
	if capacity < 2 || capacity > 10 {
		capacity = 10
	}

	r, err := h.rooms.Create(capacity)
	if err != nil {
		h.logger.Error("create room", "error", err)
		return
	}

	// Auto-join creator.
	if err := r.Join(s.PlayerID); err != nil {
		h.logger.Error("auto-join creator", "error", err)
		return
	}

	players, cap2, state, matchID := r.Snapshot()

	h.send(source, packet.PacketIDRoomCreated, protocol.RoomResponse{
		RoomID:   r.ID,
		PlayerID: s.PlayerID,
		Players:  players,
		Capacity: cap2,
		State:    string(state),
	})
	h.send(source, packet.PacketIDRoomJoined, protocol.RoomResponse{
		RoomID:   r.ID,
		PlayerID: s.PlayerID,
		Players:  players,
		Capacity: cap2,
		State:    string(state),
	})

	// Check if room is full immediately (edge case: capacity 1 → not valid per rules but guard anyway).
	_ = matchID
	if r.IsFull() {
		h.startMatch(r)
	}
}

// ── Join Room ──────────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleJoinRoom(source net.Addr, pkt packet.Packet) {
	s, ok := h.requireSession(source, pkt.ID)
	if !ok {
		return
	}

	var req protocol.JoinRoomRequest
	if err := protocol.Unmarshal(pkt.Payload, &req); err != nil || req.RoomID == "" {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidPacket, "missing room_id")
		return
	}

	r, exists := h.rooms.Get(req.RoomID)
	if !exists {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidRoom, "room not found")
		return
	}

	if err := r.Join(s.PlayerID); err != nil {
		switch {
		case errors.Is(err, room.ErrDuplicateJoin):
			h.sendError(source, uint16(pkt.ID), protocol.ErrorDuplicateJoin, "already in room")
		case errors.Is(err, room.ErrRoomFull):
			h.sendError(source, uint16(pkt.ID), protocol.ErrorRoomFull, "room full")
		default:
			h.sendError(source, uint16(pkt.ID), protocol.ErrorRoomFull, "room not joinable")
		}
		return
	}

	players, capacity, state, _ := r.Snapshot()

	joined := protocol.RoomResponse{
		RoomID:   r.ID,
		PlayerID: s.PlayerID,
		Players:  players,
		Capacity: capacity,
		State:    string(state),
	}

	// Notify joiner.
	h.send(source, packet.PacketIDRoomJoined, joined)

	// Broadcast join event to existing members.
	for _, pid := range players {
		if pid == s.PlayerID {
			continue
		}
		if ps, ok := h.sessions.GetByPlayerID(pid); ok {
			h.send(ps.Addr, packet.PacketIDRoomJoined, joined)
		}
	}

	// If room is now full, start the match.
	if r.IsFull() {
		h.startMatch(r)
	}
}

// ── Leave Room ─────────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleLeaveRoom(source net.Addr, pkt packet.Packet) {
	s, ok := h.requireSession(source, pkt.ID)
	if !ok {
		return
	}

	var req protocol.LeaveRoomRequest
	if err := protocol.Unmarshal(pkt.Payload, &req); err != nil || req.RoomID == "" {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidPacket, "missing room_id")
		return
	}

	r, exists := h.rooms.Get(req.RoomID)
	if !exists {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidRoom, "room not found")
		return
	}

	r.Leave(s.PlayerID)
	h.world.RemovePlayer(s.PlayerID)

	players, capacity, state, _ := r.Snapshot()
	left := protocol.RoomResponse{
		RoomID:   r.ID,
		PlayerID: s.PlayerID,
		Players:  players,
		Capacity: capacity,
		State:    string(state),
	}

	h.send(source, packet.PacketIDRoomLeft, left)

	// Notify remaining members.
	for _, pid := range players {
		if ps, ok := h.sessions.GetByPlayerID(pid); ok {
			h.send(ps.Addr, packet.PacketIDRoomLeft, left)
		}
	}
}

// ── Movement Input ─────────────────────────────────────────────────────────

func (h *MultiplayerHandler) handleMovementInput(source net.Addr, pkt packet.Packet) {
	s, ok := h.requireSession(source, pkt.ID)
	if !ok {
		return
	}

	var req protocol.MovementInput
	if err := protocol.Unmarshal(pkt.Payload, &req); err != nil {
		h.sendError(source, uint16(pkt.ID), protocol.ErrorInvalidPacket, "bad json")
		return
	}

	h.inputs.Add(command.Input{
		PlayerID: s.PlayerID,
		Sequence: req.Sequence,
		MoveX:    req.MoveX,
		MoveY:    req.MoveY,
		Rotation: req.Rotation,
	})
}

// ── Match Startup ──────────────────────────────────────────────────────────

// startMatch creates a Match, assigns teams, spawns entities, and notifies all players.
func (h *MultiplayerHandler) startMatch(r *room.Room) {
	players, _, _, _ := r.Snapshot()

	mt, err := h.matches.Create(r.ID)
	if err != nil {
		h.logger.Error("create match", "error", err)
		return
	}
	mt.Start()
	r.SetPlaying(mt.ID)

	// Track team slot per team for deterministic spawn indexing.
	teamSlot := map[int]int{0: 0, 1: 0}

	for _, playerID := range players {
		team := mt.AssignTeam(playerID)
		idx := teamSlot[team]
		teamSlot[team]++

		sp := game.ChooseSpawn(team, idx)
		entity := h.world.SpawnPlayer(mt.ID, r.ID, playerID, team, world.Vec2{X: sp.X, Y: sp.Y})

		spawnMsg := protocol.PlayerSpawned{
			MatchID:        mt.ID,
			RoomID:         r.ID,
			EntityID:       entity.EntityID,
			PlayerID:       playerID,
			Team:           team,
			Position:       protocol.Vec2{X: entity.Position.X, Y: entity.Position.Y},
			Rotation:       entity.Rotation,
			HP:             entity.HP,
			AnimationState: entity.AnimationState,
		}

		// Broadcast PlayerSpawned to every player in the match.
		for _, pid := range players {
			if ps, ok := h.sessions.GetByPlayerID(pid); ok {
				h.send(ps.Addr, packet.PacketIDPlayerSpawned, spawnMsg)
			}
		}
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

// requireSession looks up the session for source, sending DISCONNECTED if absent.
func (h *MultiplayerHandler) requireSession(source net.Addr, pktID packet.ID) (*session.Session, bool) {
	s, ok := h.sessions.GetByAddr(source)
	if !ok {
		h.sendError(source, uint16(pktID), protocol.ErrorDisconnected, "not logged in")
	}
	return s, ok
}

// send marshals v and sends a packet to target. Errors are logged, not returned.
func (h *MultiplayerHandler) send(target net.Addr, id packet.ID, v interface{}) {
	payload, err := protocol.Marshal(v)
	if err != nil {
		h.logger.Error("marshal packet", "id", id.String(), "error", err)
		return
	}
	if err := h.sender.Send(target, packet.Packet{ID: id, Payload: payload}); err != nil {
		h.logger.Debug("send packet", "id", id.String(), "target", target.String(), "error", err)
	}
}

// sendError sends an ErrorResponse to source.
func (h *MultiplayerHandler) sendError(source net.Addr, requestID uint16, code protocol.ErrorCode, msg string) {
	h.send(source, packet.PacketIDErrorResponse, protocol.ErrorResponse{
		RequestPacketID: requestID,
		Code:            code,
		Message:         msg,
	})
}
