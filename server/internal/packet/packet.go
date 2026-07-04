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

	// Movement / gameplay
	PacketIDMovementInput ID = 30

	// Spawn / entity
	PacketIDPlayerSpawned      ID = 31
	PacketIDPlayerDisconnected ID = 32

	// Snapshots
	PacketIDSnapshot ID = 40

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
	case PacketIDPing:
		return "Ping"
	case PacketIDPong:
		return "Pong"
	default:
		return fmt.Sprintf("Unknown(%d)", id)
	}
}
