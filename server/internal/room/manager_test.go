package room

import (
	"errors"
	"testing"
)

func TestCreateRoomStartsWaitingWithCapacity(t *testing.T) {
	manager := NewManager()

	r, err := manager.Create(2)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	players, capacity, state, matchID := r.Snapshot()
	if r.ID == "" || capacity != 2 || state != StateWaiting || matchID != "" || len(players) != 0 {
		t.Fatalf("unexpected room snapshot: id=%q players=%v capacity=%d state=%s match=%q", r.ID, players, capacity, state, matchID)
	}
}

func TestJoinRejectsDuplicateAndFullRoom(t *testing.T) {
	manager := NewManager()
	r, err := manager.Create(2)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if err := r.Join("player-1"); err != nil {
		t.Fatalf("join player 1: %v", err)
	}
	if err := r.Join("player-1"); !errors.Is(err, ErrDuplicateJoin) {
		t.Fatalf("duplicate join error = %v, want %v", err, ErrDuplicateJoin)
	}
	if err := r.Join("player-2"); err != nil {
		t.Fatalf("join player 2: %v", err)
	}
	if err := r.Join("player-3"); !errors.Is(err, ErrRoomFull) {
		t.Fatalf("full room join error = %v, want %v", err, ErrRoomFull)
	}
}

func TestPlayingRoomIsNotJoinable(t *testing.T) {
	manager := NewManager()
	r, err := manager.Create(2)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	r.SetPlaying("match-1")
	if err := r.Join("player-1"); !errors.Is(err, ErrRoomNotJoinable) {
		t.Fatalf("playing room join error = %v, want %v", err, ErrRoomNotJoinable)
	}
}

func TestLeaveRemovesPlayerAndRemoveDestroysRoom(t *testing.T) {
	manager := NewManager()
	r, err := manager.Create(2)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := r.Join("player-1"); err != nil {
		t.Fatalf("join player: %v", err)
	}

	r.Leave("player-1")
	players, _, _, _ := r.Snapshot()
	if len(players) != 0 {
		t.Fatalf("players after leave = %v, want empty", players)
	}

	manager.Remove(r.ID)
	if _, ok := manager.Get(r.ID); ok {
		t.Fatal("room still found after remove")
	}
}
