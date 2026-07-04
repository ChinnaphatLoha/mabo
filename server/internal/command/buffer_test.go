package command

import "testing"

func TestInputBufferKeepsLatestInputPerPlayer(t *testing.T) {
	buffer := NewInputBuffer()

	buffer.Add(Input{PlayerID: "player-1", Sequence: 1, MoveX: 1})
	buffer.Add(Input{PlayerID: "player-1", Sequence: 2, MoveY: 1})
	buffer.Add(Input{PlayerID: "player-2", Sequence: 1, MoveX: -1})

	inputs := buffer.Drain()
	if len(inputs) != 2 {
		t.Fatalf("drained input count = %d, want 2", len(inputs))
	}

	byPlayer := map[string]Input{}
	for _, input := range inputs {
		byPlayer[input.PlayerID] = input
	}
	if byPlayer["player-1"].Sequence != 2 || byPlayer["player-1"].MoveY != 1 {
		t.Fatalf("player-1 latest input mismatch: %#v", byPlayer["player-1"])
	}
	if byPlayer["player-2"].Sequence != 1 || byPlayer["player-2"].MoveX != -1 {
		t.Fatalf("player-2 input mismatch: %#v", byPlayer["player-2"])
	}

	if inputs := buffer.Drain(); len(inputs) != 0 {
		t.Fatalf("drain did not clear buffer: %#v", inputs)
	}
}

func TestInputBufferIgnoresOlderSequences(t *testing.T) {
	buffer := NewInputBuffer()

	buffer.Add(Input{PlayerID: "player-1", Sequence: 3, MoveX: 1})
	buffer.Add(Input{PlayerID: "player-1", Sequence: 2, MoveX: -1})

	inputs := buffer.Drain()
	if len(inputs) != 1 {
		t.Fatalf("drained input count = %d, want 1", len(inputs))
	}
	if inputs[0].Sequence != 3 || inputs[0].MoveX != 1 {
		t.Fatalf("older input replaced newer input: %#v", inputs[0])
	}
}
