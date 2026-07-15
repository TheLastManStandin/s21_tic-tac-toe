package domain

func minimax(game *minimaxGame, turn int) {
	endgame := calculateEndgame(*game)
	if endgame != Still {
		game.worstMarch = endgame
		return
	}

	games := make([]minimaxGame, 0, 9)
	oldContour := minimaxGame{}

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if game.table[y][x] == 0 {
				nextTurn := minimaxGame{
					table:      game.table,
					worstMarch: Still,
				}
				nextTurn.table[y][x] = turn
				games = append(games, nextTurn)
			}
		}
	}

	for _, contour := range games {
		oldContour = contour
		minimax(&contour, 3-turn)
		oldContour.worstMarch = contour.worstMarch

		if turn == 1 {
			if game.worstMarch == Still || contour.worstMarch < game.worstMarch {
				game.worstMarch = oldContour.worstMarch
				game.table = oldContour.table
			}
		} else {
			if contour.worstMarch > game.worstMarch {
				game.worstMarch = oldContour.worstMarch
				game.table = oldContour.table
			}
		}
	}
}
