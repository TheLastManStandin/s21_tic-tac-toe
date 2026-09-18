package domain

func (state GameState) IsFinished() bool {
	switch state {
	case FirstPlayerWon, SecondPlayerWon, Draw:
		return true
	default:
		return false
	}
}

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
