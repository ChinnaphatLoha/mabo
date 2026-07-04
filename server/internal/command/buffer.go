package command

import "sync"

type Input struct {
	PlayerID string
	Sequence uint32
	MoveX    float64
	MoveY    float64
	Rotation float64
}

type InputBuffer struct {
	mu     sync.Mutex
	inputs map[string]Input
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{inputs: make(map[string]Input)}
}

func (b *InputBuffer) Add(input Input) {
	if input.PlayerID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	current, exists := b.inputs[input.PlayerID]
	if exists && input.Sequence < current.Sequence {
		return
	}
	b.inputs[input.PlayerID] = input
}

func (b *InputBuffer) Drain() []Input {
	b.mu.Lock()
	defer b.mu.Unlock()

	inputs := make([]Input, 0, len(b.inputs))
	for _, input := range b.inputs {
		inputs = append(inputs, input)
	}
	b.inputs = make(map[string]Input)
	return inputs
}
