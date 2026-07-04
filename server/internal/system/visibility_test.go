package system

import "testing"

func TestBroadcastInterestManagerReturnsEveryPlayer(t *testing.T) {
	manager := BroadcastInterestManager{}
	recipients := manager.Recipients("viewer", []string{"player-1", "player-2"})

	if len(recipients) != 2 || recipients[0] != "player-1" || recipients[1] != "player-2" {
		t.Fatalf("recipients = %#v, want all players", recipients)
	}
}
