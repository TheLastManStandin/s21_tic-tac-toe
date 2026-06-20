package datasource

import (
	"domain"
	"sync"
)

type GamesDatabase struct {
	Games sync.Map
}

func NewGamesDatabase() *GamesDatabase {
	db := &GamesDatabase{}
	return db
}

func (gd GamesDatabase) SaveGame(game domain.Game) {
	gd.Games.Store(game.UUID, game.Table)
}

func (gd GamesDatabase) LoadGame(tableUUID string) (domain.Game, error) {
	game, ok := gd.Games.Load(tableUUID)
	if !ok {
		_ = 1 // TODO: game init
	}

	return game.(domain.Game), nil
}
