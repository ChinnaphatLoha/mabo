package world

import (
	"testing"

	"github.com/ChinnaphatLoha/mabo/server/internal/command"
)

func TestSpawnPlayerCreatesInitialStateAndSnapshot(t *testing.T) {
	w := New()

	entity := w.SpawnPlayer("match-1", "room-1", "player-1", 0, Vec2{X: 5, Y: 7})

	if entity.EntityID == "" || entity.PlayerID != "player-1" || entity.Team != 0 || entity.Position.X != 5 || entity.Position.Y != 7 {
		t.Fatalf("unexpected spawned entity: %#v", entity)
	}
	if entity.HP != 100 || entity.AnimationState != AnimationIdle {
		t.Fatalf("unexpected initial placeholders: hp=%d animation=%s", entity.HP, entity.AnimationState)
	}

	snapshot, ok := w.Snapshot("match-1", 1)
	if !ok {
		t.Fatal("snapshot not found for match")
	}
	if snapshot.Tick != 1 || len(snapshot.Players) != 1 || snapshot.Players[0].PlayerID != "player-1" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestApplyMovementInputUpdatesServerOwnedPosition(t *testing.T) {
	w := New()
	w.SpawnPlayer("match-1", "room-1", "player-1", 0, Vec2{})

	w.ApplyInputs([]command.Input{
		{PlayerID: "player-1", Sequence: 1, MoveX: 3, MoveY: 4, Rotation: 1.25},
	}, 1.0)

	snapshot, ok := w.Snapshot("match-1", 2)
	if !ok {
		t.Fatal("snapshot not found for match")
	}
	player := snapshot.Players[0]
	if player.Position.X != PlayerMoveSpeed*0.6 || player.Position.Y != PlayerMoveSpeed*0.8 {
		t.Fatalf("movement was not normalized/applied: %#v", player.Position)
	}
	if player.Rotation != 1.25 || player.AnimationState != AnimationRun {
		t.Fatalf("rotation/animation mismatch: rotation=%f animation=%s", player.Rotation, player.AnimationState)
	}
}

func TestRemovePlayerDeletesEntityFromSnapshots(t *testing.T) {
	w := New()
	w.SpawnPlayer("match-1", "room-1", "player-1", 0, Vec2{})

	removed, ok := w.RemovePlayer("player-1")
	if !ok || removed.PlayerID != "player-1" {
		t.Fatalf("remove result = %#v ok=%v", removed, ok)
	}

	snapshot, ok := w.Snapshot("match-1", 3)
	if !ok {
		t.Fatal("snapshot not found for match")
	}
	if len(snapshot.Players) != 0 {
		t.Fatalf("snapshot still has removed player: %#v", snapshot.Players)
	}
}
