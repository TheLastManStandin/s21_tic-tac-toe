package domain

type GameService interface {
	HandleTurn(game Game) (Game, error)
	ValidateGame(game Game) (error, bool)
	GameEnded() bool
}

type Service struct{}

func NewService() GameService {
	return Service{}
}
