package datasource

func NewGameRepository(db *GamesDatabase) *GameRepository {
	return &GameRepository{db: db}
}

func (gr *GameRepository) SaveGame(game DbGame) {
	gr.db.SaveGame(game)
}

func (gr *GameRepository) GetGame(uuid string) (DbGame, error) {
	game, err := gr.db.LoadGame(uuid)
	if err != nil {
		return DbGame{}, err
	}
	return game, nil
}
