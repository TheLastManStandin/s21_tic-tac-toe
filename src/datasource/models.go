package datasource

import "sync"

type Database interface {
	SaveGame(instance GameInstance, UUID string)
	LoadGame(UUID string) (GameInstance, error)
}

type GamesDatabase struct {
	Games sync.Map
}

type GameInstance struct {
	Table [3][3]int
}
