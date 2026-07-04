package session

import (
	"net"
	"time"
)

// Session represents a connected client's in-memory guest session.
type Session struct {
	ID        string
	PlayerID  string
	GuestName string
	Addr      net.Addr
	LastSeen  time.Time
}
