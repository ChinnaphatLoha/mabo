package packet

import "fmt"

type ID uint16

const (
	PacketIDConnect    ID = 1
	PacketIDDisconnect ID = 2

	// Session / Room
	PacketIDLoginRequest  ID = 10
	PacketIDLoginResponse ID = 11

	PacketIDCreateRoomRequest ID = 20
	PacketIDRoomCreated       ID = 21
	PacketIDJoinRoomRequest   ID = 22
	PacketIDRoomJoined        ID = 23
	PacketIDLeaveRoomRequest  ID = 24
	PacketIDRoomLeft          ID = 25
	PacketIDErrorResponse     ID = 49

	// Movement / gameplay
	PacketIDMovementInput ID = 150

	// Snapshots
	PacketIDSnapshot ID = 200

	// Spawn / entity
	PacketIDPlayerSpawned      ID = 250
	PacketIDPlayerDisconnected ID = 251

	// Ping
	PacketIDPing ID = 100
	PacketIDPong ID = 101
)

type Packet struct {
	ID      ID
	Payload []byte
}

func (id ID) String() string {
	switch id {
	case PacketIDConnect:
		return "Connect"
	case PacketIDDisconnect:
		return "Disconnect"
	case PacketIDLoginRequest:
		return "LoginRequest"
	case PacketIDLoginResponse:
		return "LoginResponse"
	case PacketIDCreateRoomRequest:
		return "CreateRoomRequest"
	case PacketIDRoomCreated:
		return "RoomCreated"
	case PacketIDJoinRoomRequest:
		return "JoinRoomRequest"
	case PacketIDRoomJoined:
		return "RoomJoined"
	case PacketIDLeaveRoomRequest:
		return "LeaveRoomRequest"
	case PacketIDRoomLeft:
		return "RoomLeft"
	case PacketIDErrorResponse:
		return "ErrorResponse"
	case PacketIDMovementInput:
		return "MovementInput"
	case PacketIDSnapshot:
		return "Snapshot"
	case PacketIDPlayerSpawned:
		return "PlayerSpawned"
	case PacketIDPlayerDisconnected:
		return "PlayerDisconnected"
	case PacketIDPing:
		return "Ping"
	case PacketIDPong:
		return "Pong"
	default:
		return fmt.Sprintf("Unknown(%d)", id)
	}
}
