package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	ErrInvalidRoom    = errors.New("invalid room")
	ErrRoomFull       = errors.New("room full")
	ErrDuplicateJoin  = errors.New("duplicate join")
	ErrRoomNotJoinable = errors.New("room not joinable")
)

type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func genID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) Create(capacity int) (*Room, error) {
	id, err := genID()
	if err != nil {
		return nil, err
	}
	r := &Room{ID: id, Capacity: capacity, State: StateWaiting, Players: make([]string, 0, capacity)}
	m.mu.Lock()
	m.rooms[r.ID] = r
	m.mu.Unlock()
	return r, nil
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
	if r, ok := m.rooms[id]; ok {
		r.mu.Lock()
		r.State = StateDestroyed
		r.mu.Unlock()
	}
	delete(m.rooms, id)
}

func (r *Room) Join(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateWaiting {
		return ErrRoomNotJoinable
	}
	for _, p := range r.Players {
		if p == playerID {
			return ErrDuplicateJoin
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

func (r *Room) IsFull() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Players) >= r.Capacity
}

func (r *Room) Snapshot() (players []string, capacity int, state State, matchID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	players = append([]string(nil), r.Players...)
	return players, r.Capacity, r.State, r.MatchID
}

func (r *Room) SetPlaying(matchID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.MatchID = matchID
	r.State = StatePlaying
}

func (r *Room) SetFinished() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.State = StateFinished
}

// List returns all rooms currently tracked by the manager.
func (m *Manager) List() []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		out = append(out, r)
	}
	return out
}
