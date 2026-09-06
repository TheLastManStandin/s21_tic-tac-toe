package gameService

import "tickTackToe/internal/domain"

func (gs *GameService) CalculateGameState(
	game domain.GameInstance,
	nextTurn domain.GameState,
) domain.GameState {
	for row := 0; row < 3; row++ {
		if winner := lineWinner(
			game.Table[row][0],
			game.Table[row][1],
			game.Table[row][2],
		); winner != 0 {
			return winnerState(winner)
		}
	}

	for column := 0; column < 3; column++ {
		if winner := lineWinner(
			game.Table[0][column],
			game.Table[1][column],
			game.Table[2][column],
		); winner != 0 {
			return winnerState(winner)
		}
	}

	if winner := lineWinner(
		game.Table[0][0],
		game.Table[1][1],
		game.Table[2][2],
	); winner != 0 {
		return winnerState(winner)
	}

	if winner := lineWinner(
		game.Table[0][2],
		game.Table[1][1],
		game.Table[2][0],
	); winner != 0 {
		return winnerState(winner)
	}

	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			if game.Table[row][column] == 0 {
				return nextTurn
			}
		}
	}

	return domain.Draw
}

func lineWinner(first, second, third int) int {
	if first != 0 && first == second && second == third {
		return first
	}

	return 0
}

func winnerState(winner int) domain.GameState {
	if winner == 1 {
		return domain.FirstPlayerWon
	}

	return domain.SecondPlayerWon
}
