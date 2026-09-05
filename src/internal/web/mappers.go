package web

import "tickTackToe/internal/domain"

func GameFromDomain(game domain.GameInstance) WebGame {
	return WebGame{
		Table:   game.Table,
		State:   gameStateFromDomain(game.State),
		UUID:    game.UUID,
		Minimax: game.Minimax,
	}
}

func GameToDomain(game WebGame) domain.GameInstance {
	return domain.GameInstance{
		Table:   game.Table,
		State:   gameStateToDomain(game.State),
		UUID:    game.UUID,
		Minimax: game.Minimax,
	}
}

func gameStateFromDomain(state domain.GameState) WebGameState {
	switch state {
	case domain.WaitingForPlayers:
		return WaitingForPlayers
	case domain.FirstPlayerTurn:
		return FirstPlayerTurn
	case domain.SecondPlayerTurn:
		return SecondPlayerTurn
	case domain.FirstPlayerWon:
		return FirstPlayerWon
	case domain.SecondPlayerWon:
		return SecondPlayerWon
	case domain.Draw:
		return Draw
	default:
		return WebGameState(state)
	}
}

func gameStateToDomain(state WebGameState) domain.GameState {
	switch state {
	case WaitingForPlayers:
		return domain.WaitingForPlayers
	case FirstPlayerTurn:
		return domain.FirstPlayerTurn
	case SecondPlayerTurn:
		return domain.SecondPlayerTurn
	case FirstPlayerWon:
		return domain.FirstPlayerWon
	case SecondPlayerWon:
		return domain.SecondPlayerWon
	case Draw:
		return domain.Draw
	default:
		return domain.GameState(state)
	}
}
