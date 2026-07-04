package room

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type State int

const (
	StateCreated State = iota
	StateWaiting
	StatePlaying
	StateFinished
	StateDestroyed
)

type Room struct {
	ID       string
	Players  []string // player IDs
	Capacity int
	State    State
	mu       sync.Mutex
}

type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func genID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) Create(capacity int) *Room {
	r := &Room{ID: genID(), Capacity: capacity, State: StateWaiting, Players: make([]string, 0)}
	m.mu.Lock()
	m.rooms[r.ID] = r
	m.mu.Unlock()
	return r
}

func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	return r, ok
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (r *Room) Join(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// check duplicate
	for _, p := range r.Players {
		if p == playerID {
			return nil
		}
	}
	if len(r.Players) >= r.Capacity {
		return ErrRoomFull
	}
	r.Players = append(r.Players, playerID)
	return nil
}

func (r *Room) Leave(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, p := range r.Players {
		if p == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			break
		}
	}
}

var ErrRoomFull = &RoomError{"room full"}

type RoomError struct{ s string }

func (e *RoomError) Error() string { return e.s }
