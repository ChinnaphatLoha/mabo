package command

import "testing"

func TestInputBufferQueuesInputsPerPlayer(t *testing.T) {
	buffer := NewInputBuffer()

	buffer.Add(Input{PlayerID: "player-1", Sequence: 2, MoveY: 1})
	buffer.Add(Input{PlayerID: "player-1", Sequence: 1, MoveX: 1}) // Out of order insert
	buffer.Add(Input{PlayerID: "player-2", Sequence: 1, MoveX: -1})
	buffer.Add(Input{PlayerID: "player-1", Sequence: 2, MoveY: 1}) // Duplicate insert

	inputs := buffer.Drain()
	if len(inputs) != 3 {
		t.Fatalf("drained input count = %d, want 3", len(inputs))
	}

	byPlayer := map[string][]Input{}
	for _, input := range inputs {
		byPlayer[input.PlayerID] = append(byPlayer[input.PlayerID], input)
	}

	p1Inputs := byPlayer["player-1"]
	if len(p1Inputs) != 2 {
		t.Fatalf("player-1 should have 2 unique inputs, got %d", len(p1Inputs))
	}
	if p1Inputs[0].Sequence != 1 || p1Inputs[1].Sequence != 2 {
		t.Fatalf("player-1 inputs not sorted correctly: %v", p1Inputs)
	}

	if byPlayer["player-2"][0].Sequence != 1 {
		t.Fatalf("player-2 input mismatch")
	}

	if inputs := buffer.Drain(); len(inputs) != 0 {
		t.Fatalf("drain did not clear buffer: %#v", inputs)
	}
}

