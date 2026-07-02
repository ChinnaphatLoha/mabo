package packet

import "fmt"

type ID uint16

const (
	PacketIDConnect    ID = 1
	PacketIDDisconnect ID = 2
	PacketIDPing       ID = 100
	PacketIDPong       ID = 101
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
