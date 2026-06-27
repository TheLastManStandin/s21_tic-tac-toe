package datasource

import (
	"sync"
)

type GamesDatabase struct {
	Games sync.Map
}

type GameRepository struct {
	db *GamesDatabase
}
