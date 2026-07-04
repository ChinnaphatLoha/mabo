package packet

import "testing"

func TestDay2PacketIDsFollowProtocolRanges(t *testing.T) {
	tests := map[ID]ID{
		PacketIDLoginRequest:       10,
		PacketIDLoginResponse:      11,
		PacketIDCreateRoomRequest:  20,
		PacketIDRoomCreated:        21,
		PacketIDJoinRoomRequest:    22,
		PacketIDRoomJoined:         23,
		PacketIDLeaveRoomRequest:   24,
		PacketIDRoomLeft:           25,
		PacketIDErrorResponse:      49,
		PacketIDMovementInput:      150,
		PacketIDSnapshot:           200,
		PacketIDPlayerSpawned:      250,
		PacketIDPlayerDisconnected: 251,
	}

	for got, want := range tests {
		if got != want {
			t.Fatalf("packet ID mismatch: got %d want %d", got, want)
		}
	}
}

func TestPacketIDStringIncludesDay2Packets(t *testing.T) {
	tests := map[ID]string{
		PacketIDLoginRequest:       "LoginRequest",
		PacketIDLoginResponse:      "LoginResponse",
		PacketIDCreateRoomRequest:  "CreateRoomRequest",
		PacketIDRoomCreated:        "RoomCreated",
		PacketIDJoinRoomRequest:    "JoinRoomRequest",
		PacketIDRoomJoined:         "RoomJoined",
		PacketIDLeaveRoomRequest:   "LeaveRoomRequest",
		PacketIDRoomLeft:           "RoomLeft",
		PacketIDErrorResponse:      "ErrorResponse",
		PacketIDMovementInput:      "MovementInput",
		PacketIDSnapshot:           "Snapshot",
		PacketIDPlayerSpawned:      "PlayerSpawned",
		PacketIDPlayerDisconnected: "PlayerDisconnected",
	}

	for id, want := range tests {
		if got := id.String(); got != want {
			t.Fatalf("packet %d string mismatch: got %q want %q", id, got, want)
		}
	}
}
