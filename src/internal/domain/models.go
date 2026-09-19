package domain

type GameInstance struct {
	Table   [3][3]int
	State   GameState
	UUID    string
	Minimax bool
}

type GameState int32

const (
	WaitingForPlayers GameState = iota
	FirstPlayerTurn
	SecondPlayerTurn
	FirstPlayerWon
	SecondPlayerWon
	Draw
)

type UserInstance struct {
	Username string
	Password string
	UUID     string
}
