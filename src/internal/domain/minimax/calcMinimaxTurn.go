package minimax

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

func calculateEndgame(game minimaxGame) MinimaxStatus {
	for i := 0; i < 3; i++ {
		rowVal := game.table[i][0]
		rowSame := true
		for j := 0; j < 3; j++ {
			if game.table[i][j] != rowVal {
				rowSame = false
			}
		}

		if rowSame && rowVal != 0 {
			if rowVal == 1 {
				return MinimaxLose
			} else if rowVal == 2 {
				return MinimaxWin
			}
		}
	}

	for i := 0; i < 3; i++ {
		columnVal := game.table[0][i]
		columnSame := true
		for j := 0; j < 3; j++ {
			if game.table[j][i] != columnVal {
				columnSame = false
			}
		}

		if columnSame && columnVal != 0 {
			if columnVal == 1 {
				return MinimaxLose
			} else if columnVal == 2 {
				return MinimaxWin
			}
		}
	}

	midVal := game.table[1][1]
	if game.table[0][0] == midVal && game.table[2][2] == midVal && midVal != 0 {
		if midVal == 1 {
			return MinimaxLose
		} else if midVal == 2 {
			return MinimaxWin
		}
	} else if game.table[0][2] == midVal && game.table[2][0] == midVal && midVal != 0 {
		if midVal == 1 {
			return MinimaxLose
		} else if midVal == 2 {
			return MinimaxWin
		}
	}

	haveSpace := false
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if game.table[i][j] == 0 {
				haveSpace = true
				break
			}
		}
	}
	if haveSpace {
		return Still
	} else {
		return Draw
	}
}
