package datasource

import "domain"

func NewGameRepository(db *GamesDatabase) *GameRepository {
	return &GameRepository{db: db}
}

func (gr *GameRepository) SaveGame(game domain.Game) {
	gr.db.SaveGame(game)
}

func (gr *GameRepository) GetGame(uuid string) (domain.Game, error) {
	game, err := gr.db.LoadGame(uuid)
	if err != nil {
		return domain.Game{}, err
	}
	return game, nil
}
