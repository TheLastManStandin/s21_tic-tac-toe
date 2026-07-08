package datasource

func NewGamesDatabase() *GamesDatabase {
	db := &GamesDatabase{}
	return db
}

func (gd *GamesDatabase) SaveGame(game DbGame) {
	gd.Games.Store(game.UUID, game)
}

func (gd *GamesDatabase) LoadGame(tableUUID string) (DbGame, error) {
	game, ok := gd.Games.Load(tableUUID)
	if !ok {
		//_ = 1 // TODO: game init
		return DbGame{}, nil
	}

	return game.(DbGame), nil
}
