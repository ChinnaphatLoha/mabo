package command

import (
	"sort"
	"sync"
)

type Input struct {
	PlayerID string
	Sequence uint32
	MoveX    float64
	MoveY    float64
	Rotation float64
}

type InputBuffer struct {
	mu     sync.Mutex
	inputs map[string][]Input
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{inputs: make(map[string][]Input)}
}

func (b *InputBuffer) Add(input Input) {
	if input.PlayerID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Append to player's queue
	b.inputs[input.PlayerID] = append(b.inputs[input.PlayerID], input)
}

func (b *InputBuffer) Drain() []Input {
	b.mu.Lock()
	defer b.mu.Unlock()

	var allInputs []Input
	for playerID, playerInputs := range b.inputs {
		// Sort by sequence to ensure they are processed in order
		sort.Slice(playerInputs, func(i, j int) bool {
			return playerInputs[i].Sequence < playerInputs[j].Sequence
		})
		
		// Remove duplicates
		if len(playerInputs) > 0 {
			uniqueInputs := []Input{playerInputs[0]}
			for i := 1; i < len(playerInputs); i++ {
				if playerInputs[i].Sequence != uniqueInputs[len(uniqueInputs)-1].Sequence {
					uniqueInputs = append(uniqueInputs, playerInputs[i])
				}
			}
			allInputs = append(allInputs, uniqueInputs...)
		}
		
		// Clear player's queue after draining
		delete(b.inputs, playerID)
	}

	return allInputs
}
