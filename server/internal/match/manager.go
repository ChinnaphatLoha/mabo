package match

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type State string

const (
	StateCreated   State = "created"
	StateRunning   State = "running"
	StateFinished  State = "finished"
	StateDestroyed State = "destroyed"
)

type Match struct {
	ID     string
	RoomID string
	Teams  map[string]int // playerID -> team (0 or 1)
	State  State
	mu     sync.Mutex
}

type Manager struct {
	mu            sync.RWMutex
	matches       map[string]*Match
	matchesByRoom map[string]*Match
}

func NewManager() *Manager {
	return &Manager{
		matches:       make(map[string]*Match),
		matchesByRoom: make(map[string]*Match),
	}
}

func genID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) Create(roomID string) (*Match, error) {
	id, err := genID()
	if err != nil {
		return nil, err
	}
	mt := &Match{ID: id, RoomID: roomID, Teams: make(map[string]int), State: StateCreated}
	m.mu.Lock()
	m.matches[mt.ID] = mt
	m.matchesByRoom[roomID] = mt
	m.mu.Unlock()
	return mt, nil
}

func (m *Manager) Get(id string) (*Match, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mt, ok := m.matches[id]
	return mt, ok
}

func (m *Manager) GetByRoomID(roomID string) (*Match, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mt, ok := m.matchesByRoom[roomID]
	return mt, ok
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if mt, ok := m.matches[id]; ok {
		mt.mu.Lock()
		mt.State = StateDestroyed
		mt.mu.Unlock()
		delete(m.matchesByRoom, mt.RoomID)
	}
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

func (mt *Match) Start() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.State = StateRunning
}

func (mt *Match) Finish() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.State = StateFinished
}

func (mt *Match) PlayerIDs() []string {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	players := make([]string, 0, len(mt.Teams))
	for playerID := range mt.Teams {
		players = append(players, playerID)
	}
	return players
}

func (mt *Match) RemovePlayer(playerID string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	delete(mt.Teams, playerID)
}
