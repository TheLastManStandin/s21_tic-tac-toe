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

func (state GameState) String() string {
	switch state {
	case WaitingForPlayers:
		return "waiting_for_players"
	case FirstPlayerTurn:
		return "first_player_turn"
	case SecondPlayerTurn:
		return "second_player_turn"
	case FirstPlayerWon:
		return "first_player_won"
	case SecondPlayerWon:
		return "second_player_won"
	case Draw:
		return "draw"
	default:
		return "unknown"
	}
}
