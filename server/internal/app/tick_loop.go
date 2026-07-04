package app

import (
	"context"
	"time"

	"github.com/ChinnaphatLoha/mabo/server/internal/command"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/match"
	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
	"github.com/ChinnaphatLoha/mabo/server/internal/protocol"
	"github.com/ChinnaphatLoha/mabo/server/internal/session"
	"github.com/ChinnaphatLoha/mabo/server/internal/system"
	"github.com/ChinnaphatLoha/mabo/server/internal/world"
)

const (
	TickRate     = 20
	tickDelta    = float64(1) / float64(TickRate)
	tickInterval = time.Second / TickRate
)

// TickLoop runs the authoritative simulation at 20 TPS and broadcasts
// world snapshots to every player in each running match.
type TickLoop struct {
	sessions *session.Manager
	matches  *match.Manager
	world    *world.World
	inputs   *command.InputBuffer
	interest system.InterestManager
	sender   Sender
	logger   *logger.Logger
}

// NewTickLoop creates a tick loop with the shared managers.
// inputs must be the same *command.InputBuffer that the MultiplayerHandler writes into.
func NewTickLoop(
	sessions *session.Manager,
	matches *match.Manager,
	w *world.World,
	inputs *command.InputBuffer,
	interest system.InterestManager,
	sender Sender,
	log *logger.Logger,
) *TickLoop {
	return &TickLoop{
		sessions: sessions,
		matches:  matches,
		world:    w,
		inputs:   inputs,
		interest: interest,
		sender:   sender,
		logger:   log,
	}
}

// Run starts the tick loop and blocks until ctx is cancelled.
func (tl *TickLoop) Run(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	var tick uint64

	for {
		select {
		case <-ctx.Done():
			tl.logger.Info("tick loop stopping")
			return
		case <-ticker.C:
			tick++
			tl.tick(tick)
		}
	}
}

func (tl *TickLoop) tick(tick uint64) {
	// Drain all buffered inputs and apply them to the world.
	inputs := tl.inputs.Drain()
	tl.world.ApplyInputs(inputs, tickDelta)

	// Produce one snapshot per running match and broadcast to recipients.
	snapshots := tl.world.Snapshots(tick)
	for _, snap := range snapshots {
		tl.broadcastSnapshot(snap)
	}
}

// ExposedTick runs a single simulation tick. Exported for use in tests.
func (tl *TickLoop) ExposedTick(tick uint64) {
	tl.tick(tick)
}

func (tl *TickLoop) broadcastSnapshot(snap world.Snapshot) {
	mt, ok := tl.matches.Get(snap.MatchID)
	if !ok {
		return
	}
	if mt.State != match.StateRunning {
		return
	}

	// Build player ID list for interest evaluation.
	playerIDs := mt.PlayerIDs()

	// Convert world snapshot to protocol DTO.
	protoPlayers := make([]protocol.SnapshotPlayer, len(snap.Players))
	for i, p := range snap.Players {
		protoPlayers[i] = protocol.SnapshotPlayer{
			PlayerID:       p.PlayerID,
			EntityID:       p.EntityID,
			Team:           p.Team,
			Position:       protocol.Vec2{X: p.Position.X, Y: p.Position.Y},
			Rotation:       p.Rotation,
			Velocity:       protocol.Vec2{X: p.Velocity.X, Y: p.Velocity.Y},
			AnimationState: p.AnimationState,
			HP:             p.HP,
		}
	}

	msg := protocol.Snapshot{
		Tick:    snap.Tick,
		MatchID: snap.MatchID,
		Players: protoPlayers,
	}

	payload, err := protocol.Marshal(msg)
	if err != nil {
		tl.logger.Error("marshal snapshot", "error", err)
		return
	}

	pkt := packet.Packet{ID: packet.PacketIDSnapshot, Payload: payload}

	// For each player, ask the interest manager who should receive the snapshot.
	// BroadcastInterestManager returns every player; future managers can filter.
	sent := map[string]bool{}
	for _, viewerID := range playerIDs {
		recipients := tl.interest.Recipients(viewerID, playerIDs)
		for _, recipientID := range recipients {
			if sent[recipientID] {
				continue
			}
			sent[recipientID] = true
			if s, ok := tl.sessions.GetByPlayerID(recipientID); ok {
				if err := tl.sender.Send(s.Addr, pkt); err != nil {
					tl.logger.Debug("snapshot send", "player", recipientID, "error", err)
				}
			}
		}
	}
}
