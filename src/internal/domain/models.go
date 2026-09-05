package domain

type GameInstance struct {
	Table [3][3]int
	State GameState
	UUID  string
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
