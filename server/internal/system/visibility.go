package system

// InterestManager decides which players receive a given snapshot or event.
// The viewer argument is the observer's player ID; players is the full set.
type InterestManager interface {
	Recipients(viewer string, players []string) []string
}

// BroadcastInterestManager sends every snapshot to every player in the match.
// This is the Day 2 implementation; a spatial/distance filter can replace it
// later without touching room, match, or world logic.
type BroadcastInterestManager struct{}

// Recipients returns the full player list unchanged.
func (BroadcastInterestManager) Recipients(_ string, players []string) []string {
	return players
}
