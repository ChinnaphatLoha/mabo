package world

import (
	"math"
	"sort"
	"sync"

	"github.com/ChinnaphatLoha/mabo/server/internal/command"
	"github.com/ChinnaphatLoha/mabo/server/internal/game"
)

const (
	PlayerMoveSpeed = 5.0
	AnimationIdle   = "idle"
	AnimationRun    = "run"
)

type Vec2 struct {
	X float64
	Y float64
}

type PlayerEntity struct {
	EntityID       string
	PlayerID       string
	MatchID        string
	RoomID         string
	Team           int
	Position       Vec2
	Rotation              float64
	Velocity              Vec2
	AnimationState        string
	HP                    int
	LastProcessedSequence uint32
}

type Snapshot struct {
	Tick    uint64
	MatchID string
	Players []SnapshotPlayer
}

type SnapshotPlayer struct {
	PlayerID       string
	EntityID       string
	Team           int
	Position       Vec2
	Rotation              float64
	Velocity              Vec2
	AnimationState        string
	HP                    int
	LastProcessedSequence uint32
}

type World struct {
	mu          sync.RWMutex
	matches     map[string]map[string]*PlayerEntity
	playerMatch map[string]string
}

func New() *World {
	return &World{
		matches:     make(map[string]map[string]*PlayerEntity),
		playerMatch: make(map[string]string),
	}
}

func (w *World) SpawnPlayer(matchID, roomID, playerID string, team int, position Vec2) PlayerEntity {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.matches[matchID]; !ok {
		w.matches[matchID] = make(map[string]*PlayerEntity)
	}

	entity := &PlayerEntity{
		EntityID:       game.SpawnEntity("player", playerID),
		PlayerID:       playerID,
		MatchID:        matchID,
		RoomID:         roomID,
		Team:           team,
		Position:       position,
		AnimationState: AnimationIdle,
		HP:             100,
	}

	w.matches[matchID][playerID] = entity
	w.playerMatch[playerID] = matchID

	return cloneEntity(entity)
}

func (w *World) ApplyInputs(inputs []command.Input, deltaSeconds float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, input := range inputs {
		matchID, ok := w.playerMatch[input.PlayerID]
		if !ok {
			continue
		}
		entity, ok := w.matches[matchID][input.PlayerID]
		if !ok {
			continue
		}

		// Discard inputs that have already been processed or are out of order
		if input.Sequence <= entity.LastProcessedSequence {
			continue
		}
		
		dirX, dirY := normalized(input.MoveX, input.MoveY)
		entity.Velocity = Vec2{X: dirX * PlayerMoveSpeed, Y: dirY * PlayerMoveSpeed}
		entity.Position.X += entity.Velocity.X * deltaSeconds
		entity.Position.Y += entity.Velocity.Y * deltaSeconds
		entity.Rotation = input.Rotation
		if entity.Velocity.X == 0 && entity.Velocity.Y == 0 {
			entity.AnimationState = AnimationIdle
		} else {
			entity.AnimationState = AnimationRun
		}
		
		entity.LastProcessedSequence = input.Sequence
	}
}

func (w *World) Snapshot(matchID string, tick uint64) (Snapshot, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entities, ok := w.matches[matchID]
	if !ok {
		return Snapshot{}, false
	}

	players := make([]SnapshotPlayer, 0, len(entities))
	for _, entity := range entities {
		players = append(players, SnapshotPlayer{
			PlayerID:       entity.PlayerID,
			EntityID:       entity.EntityID,
			Team:           entity.Team,
			Position:       entity.Position,
			Rotation:       entity.Rotation,
			Velocity:              entity.Velocity,
			AnimationState:        entity.AnimationState,
			HP:                    entity.HP,
			LastProcessedSequence: entity.LastProcessedSequence,
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].PlayerID < players[j].PlayerID
	})

	return Snapshot{Tick: tick, MatchID: matchID, Players: players}, true
}

func (w *World) Snapshots(tick uint64) []Snapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	matchIDs := make([]string, 0, len(w.matches))
	for matchID := range w.matches {
		matchIDs = append(matchIDs, matchID)
	}
	sort.Strings(matchIDs)

	snapshots := make([]Snapshot, 0, len(matchIDs))
	for _, matchID := range matchIDs {
		players := make([]SnapshotPlayer, 0, len(w.matches[matchID]))
		for _, entity := range w.matches[matchID] {
			players = append(players, SnapshotPlayer{
				PlayerID:       entity.PlayerID,
				EntityID:       entity.EntityID,
				Team:           entity.Team,
				Position:       entity.Position,
				Rotation:       entity.Rotation,
				Velocity:              entity.Velocity,
				AnimationState:        entity.AnimationState,
				HP:                    entity.HP,
				LastProcessedSequence: entity.LastProcessedSequence,
			})
		}
		sort.Slice(players, func(i, j int) bool {
			return players[i].PlayerID < players[j].PlayerID
		})
		snapshots = append(snapshots, Snapshot{Tick: tick, MatchID: matchID, Players: players})
	}
	return snapshots
}

func (w *World) RemovePlayer(playerID string) (PlayerEntity, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	matchID, ok := w.playerMatch[playerID]
	if !ok {
		return PlayerEntity{}, false
	}
	entity, ok := w.matches[matchID][playerID]
	if !ok {
		delete(w.playerMatch, playerID)
		return PlayerEntity{}, false
	}

	removed := cloneEntity(entity)
	delete(w.matches[matchID], playerID)
	delete(w.playerMatch, playerID)
	return removed, true
}

func (w *World) RemoveMatch(matchID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for playerID := range w.matches[matchID] {
		delete(w.playerMatch, playerID)
	}
	delete(w.matches, matchID)
}

func normalized(x, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func cloneEntity(entity *PlayerEntity) PlayerEntity {
	if entity == nil {
		return PlayerEntity{}
	}
	return *entity
}
