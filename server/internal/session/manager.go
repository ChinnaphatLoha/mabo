package session

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
	"time"
)

// Session represents a connected client's session.
type Session struct {
	ID        string
	PlayerID  string
	Addr      net.Addr
	LastSeen  time.Time
	PositionX float64
	PositionY float64
}

// Manager holds sessions in-memory.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session // key: addr.String()
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

func genID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) Create(addr net.Addr) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := addr.String()
	s := &Session{
		ID:       genID(),
		PlayerID: "player-" + genID(),
		Addr:     addr,
		LastSeen: time.Now(),
	}
	m.sessions[key] = s
	return s
}

func (m *Manager) GetByAddr(addr net.Addr) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[addr.String()]
	return s, ok
}

func (m *Manager) RemoveByAddr(addr net.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, addr.String())
}

func (m *Manager) UpdateLastSeen(addr net.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[addr.String()]; ok {
		s.LastSeen = time.Now()
	}
}

func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s)
	}
	return out
}
