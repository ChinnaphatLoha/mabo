package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"sync"
	"time"
)

var ErrAlreadyConnected = errors.New("already connected")

// Manager holds sessions in-memory.
type Manager struct {
	mu             sync.RWMutex
	sessionsByAddr map[string]*Session
	playerIndex    map[string]*Session
}

func NewManager() *Manager {
	return &Manager{
		sessionsByAddr: make(map[string]*Session),
		playerIndex:    make(map[string]*Session),
	}
}

func genID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

func (m *Manager) Create(addr net.Addr, guestName string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := addr.String()
	if _, exists := m.sessionsByAddr[key]; exists {
		return nil, ErrAlreadyConnected
	}

	sessionID, err := genID("session-")
	if err != nil {
		return nil, err
	}
	playerID, err := genID("player-")
	if err != nil {
		return nil, err
	}

	s := &Session{
		ID:        sessionID,
		PlayerID:  playerID,
		GuestName: guestName,
		Addr:      addr,
		LastSeen:  time.Now(),
	}
	m.sessionsByAddr[key] = s
	m.playerIndex[playerID] = s
	return s, nil
}

func (m *Manager) GetByAddr(addr net.Addr) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessionsByAddr[addr.String()]
	return s, ok
}

func (m *Manager) GetByPlayerID(playerID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.playerIndex[playerID]
	return s, ok
}

func (m *Manager) RemoveByAddr(addr net.Addr) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := addr.String()
	s, ok := m.sessionsByAddr[key]
	if !ok {
		return nil, false
	}
	delete(m.sessionsByAddr, key)
	delete(m.playerIndex, s.PlayerID)
	return s, true
}

func (m *Manager) UpdateLastSeen(addr net.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessionsByAddr[addr.String()]; ok {
		s.LastSeen = time.Now()
	}
}

func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessionsByAddr))
	for _, s := range m.sessionsByAddr {
		out = append(out, s)
	}
	return out
}
