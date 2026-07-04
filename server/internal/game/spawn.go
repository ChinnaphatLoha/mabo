package game

import "fmt"

// SpawnPoint is a simple 2D position.
type SpawnPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SpawnEntity returns a deterministic player entity ID.
func SpawnEntity(entityType string, playerID string) string {
	return fmt.Sprintf("%s-%s", entityType, playerID)
}

// ChooseSpawn chooses a deterministic spawn point for a team slot.
func ChooseSpawn(team int, index int) SpawnPoint {
	if team == 0 {
		return SpawnPoint{X: float64(index * 2), Y: 0}
	}
	return SpawnPoint{X: 100 + float64(index*2), Y: 0}
}
