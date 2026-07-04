package match

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Match struct {
	ID     string
	RoomID string
	Teams  map[string]int // playerID -> team (0 or 1)
	mu     sync.Mutex
}

type Manager struct {
	mu      sync.RWMutex
	matches map[string]*Match
}

func NewManager() *Manager {
	return &Manager{matches: make(map[string]*Match)}
}

func genID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) Create(roomID string) *Match {
	mt := &Match{ID: genID(), RoomID: roomID, Teams: make(map[string]int)}
	m.mu.Lock()
	m.matches[mt.ID] = mt
	m.mu.Unlock()
	return mt
}

func (m *Manager) Get(id string) (*Match, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mt, ok := m.matches[id]
	return mt, ok
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.matches, id)
}

func (mt *Match) AssignTeam(playerID string) int {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	// simple team assignment: balance by count
	c0, c1 := 0, 0
	for _, t := range mt.Teams {
		if t == 0 {
			c0++
		} else {
			c1++
		}
	}
	team := 0
	if c1 < c0 {
		team = 1
	}
	mt.Teams[playerID] = team
	return team
}

func (mt *Match) RemovePlayer(playerID string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	delete(mt.Teams, playerID)
}
