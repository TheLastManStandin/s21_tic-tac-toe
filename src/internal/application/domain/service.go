package domain

import "datasource"

type GameService interface {
	HandleTurn(game Game) (Game, error)
	ValidateGameTurn(savedGame Game, recivedGame Game) error
	GameEnded(game Game) int
}

type Service struct {
	db *datasource.GameRepository
}

func NewService(db *datasource.GameRepository) *Service {
	return &Service{
		db: db,
	}
}
