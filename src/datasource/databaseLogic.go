package datasource

import "errors"

func NewGamesDatabase() *GamesDatabase {
	db := &GamesDatabase{}
	return db
}

func (gd *GamesDatabase) SaveGame(game GameInstance, UUID string) {
	gd.Games.Store(UUID, game)
}

func (gd *GamesDatabase) LoadGame(UUID string) (GameInstance, error) {
	game, ok := gd.Games.Load(UUID)
	if !ok {
		return GameInstance{}, errors.New("game not found")
	}

	return game.(GameInstance), nil
}
