package domain

type GameService interface {
	MinimaxNextTurn()
	ValidateGame() bool
	GameEnded() bool
}
