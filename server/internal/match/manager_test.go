package match

import "testing"

func TestCreateMatchTracksRoomAndLifecycle(t *testing.T) {
	manager := NewManager()

	mt, err := manager.Create("room-1")
	if err != nil {
		t.Fatalf("create match: %v", err)
	}

	if mt.ID == "" || mt.RoomID != "room-1" || mt.State != StateCreated {
		t.Fatalf("unexpected match: %#v", mt)
	}

	byRoom, ok := manager.GetByRoomID("room-1")
	if !ok || byRoom.ID != mt.ID {
		t.Fatalf("match not found by room: ok=%v match=%#v", ok, byRoom)
	}

	mt.Start()
	if mt.State != StateRunning {
		t.Fatalf("state after start = %s, want %s", mt.State, StateRunning)
	}

	mt.Finish()
	if mt.State != StateFinished {
		t.Fatalf("state after finish = %s, want %s", mt.State, StateFinished)
	}
}

func TestAssignTeamBalancesPlayers(t *testing.T) {
	manager := NewManager()
	mt, err := manager.Create("room-1")
	if err != nil {
		t.Fatalf("create match: %v", err)
	}

	teams := []int{
		mt.AssignTeam("player-1"),
		mt.AssignTeam("player-2"),
		mt.AssignTeam("player-3"),
		mt.AssignTeam("player-4"),
	}

	teamCounts := map[int]int{}
	for _, team := range teams {
		teamCounts[team]++
	}
	if teamCounts[0] != 2 || teamCounts[1] != 2 {
		t.Fatalf("team counts = %#v, want 2 per team", teamCounts)
	}
}

func TestRemoveMatchDeletesRoomIndex(t *testing.T) {
	manager := NewManager()
	mt, err := manager.Create("room-1")
	if err != nil {
		t.Fatalf("create match: %v", err)
	}

	manager.Remove(mt.ID)

	if _, ok := manager.Get(mt.ID); ok {
		t.Fatal("match still found by id after remove")
	}
	if _, ok := manager.GetByRoomID("room-1"); ok {
		t.Fatal("match still found by room after remove")
	}
}
