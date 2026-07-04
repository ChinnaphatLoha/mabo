package room

import "sync"

type State string

const (
	StateCreated   State = "created"
	StateWaiting   State = "waiting"
	StatePlaying   State = "playing"
	StateFinished  State = "finished"
	StateDestroyed State = "destroyed"
)

// Room represents lobby membership and lifecycle. Match simulation is tracked
// separately so matchmaking, replay, and running-world state can evolve without
// changing lobby semantics.
type Room struct {
	ID       string
	Players  []string
	Capacity int
	State    State
	MatchID  string
	mu       sync.Mutex
}
