package protocol

import "testing"

func TestRoomResponseRoundTrip(t *testing.T) {
	payload := RoomResponse{
		RoomID:   "room-1",
		PlayerID: "player-1",
		Players:  []string{"player-1", "player-2"},
		Capacity: 2,
		State:    "playing",
	}

	data, err := Marshal(payload)
	if err != nil {
		t.Fatalf("marshal room response: %v", err)
	}

	var got RoomResponse
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal room response: %v", err)
	}

	if got.RoomID != payload.RoomID || got.PlayerID != payload.PlayerID || got.Capacity != payload.Capacity || got.State != payload.State {
		t.Fatalf("room response mismatch: got %#v want %#v", got, payload)
	}
	if len(got.Players) != 2 || got.Players[0] != "player-1" || got.Players[1] != "player-2" {
		t.Fatalf("players mismatch: got %#v", got.Players)
	}
}

func TestSnapshotRoundTripIncludesFutureFields(t *testing.T) {
	payload := Snapshot{
		Tick:    7,
		MatchID: "match-1",
		Players: []SnapshotPlayer{
			{
				PlayerID:       "player-1",
				EntityID:       "entity-1",
				Team:           0,
				Position:       Vec2{X: 3, Y: 4},
				Rotation:       1.5,
				Velocity:       Vec2{X: 1, Y: 0},
				AnimationState: "run",
				HP:             100,
			},
		},
	}

	data, err := Marshal(payload)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	var got Snapshot
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	if got.Tick != 7 || got.MatchID != "match-1" || len(got.Players) != 1 {
		t.Fatalf("snapshot header mismatch: got %#v", got)
	}
	player := got.Players[0]
	if player.PlayerID != "player-1" || player.Position.X != 3 || player.Position.Y != 4 || player.Velocity.X != 1 || player.AnimationState != "run" || player.HP != 100 {
		t.Fatalf("snapshot player mismatch: got %#v", player)
	}
}

func TestErrorResponseRoundTrip(t *testing.T) {
	payload := ErrorResponse{
		RequestPacketID: 22,
		Code:            ErrorInvalidRoom,
		Message:         "invalid room",
	}

	data, err := Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error response: %v", err)
	}

	var got ErrorResponse
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}

	if got.RequestPacketID != payload.RequestPacketID || got.Code != ErrorInvalidRoom || got.Message != payload.Message {
		t.Fatalf("error response mismatch: got %#v want %#v", got, payload)
	}
}
