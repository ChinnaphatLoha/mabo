package game

import "fmt"

// SpawnPoint is a simple 2D position.
type SpawnPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SpawnEntity returns a simple generated entity ID for a spawn.
func SpawnEntity(entityType string, pt SpawnPoint) string {
	return fmt.Sprintf("%s-%.0f-%.0f", entityType, pt.X, pt.Y)
}

// ChooseSpawn chooses a spawn point for a team.
func ChooseSpawn(team int) SpawnPoint {
	if team == 0 {
		return SpawnPoint{X: 0, Y: 0}
	}
	return SpawnPoint{X: 10, Y: 0}
}
