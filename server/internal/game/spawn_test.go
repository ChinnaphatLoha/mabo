package game

import "testing"

func TestChooseSpawnIsDeterministicByTeamAndIndex(t *testing.T) {
	first := ChooseSpawn(0, 0)
	again := ChooseSpawn(0, 0)
	second := ChooseSpawn(0, 1)
	enemy := ChooseSpawn(1, 0)

	if first != again {
		t.Fatalf("same team/index returned different spawn points: %#v %#v", first, again)
	}
	if first == second {
		t.Fatalf("same team different index returned same spawn: %#v", first)
	}
	if first == enemy {
		t.Fatalf("different teams returned same spawn: %#v", first)
	}
}

func TestSpawnEntityIDIncludesPlayerID(t *testing.T) {
	id := SpawnEntity("player", "player-1")
	if id != "player-player-1" {
		t.Fatalf("entity id = %q, want %q", id, "player-player-1")
	}
}
