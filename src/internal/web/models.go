package web

type WebGame struct {
	Table   [3][3]int
	State   WebGameState
	UUID    string
	Minimax bool
}

type WebGameState int32

const (
	WaitingForPlayers WebGameState = iota
	FirstPlayerTurn
	SecondPlayerTurn
	FirstPlayerWon
	SecondPlayerWon
	Draw
)
