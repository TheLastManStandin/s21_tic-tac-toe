package AI

func Minimax(game MinimaxGame, turn int) MinimaxGame {
	// for debug =========================
	//{
	//	debugTable := [3][3]int8{
	//		{2, 2, 1},
	//		{0, 1, 0},
	//		{0, 0, 1},
	//	}
	//
	//	if reflect.DeepEqual(debugTable, game.Table) {
	//		println("debug table matched")
	//	}
	//}
	// ===================================
	if turn == 2 { // хотя бы один путь
		moddedGame := MinimaxGame{
			Table: game.Table,
			State: 0,
		}

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {

				if moddedGame.Table[i][j] == 0 {
					moddedGame.Table[i][j] = 2

					endgame := CalculateEndgame(moddedGame)
					if endgame == 2 || endgame == -1 {
						//moddedGame.Table[i][j] = 0
						moddedGame.State = 1
						return moddedGame
					}
					newGame := Minimax(moddedGame, 1)
					if newGame.State == 1 {
						//moddedGame.Table[i][j] = 0
						moddedGame.State = 1
						return moddedGame
					}

					moddedGame.Table[i][j] = 0
				}
			}
		}
		moddedGame.State = 0
		return moddedGame
	} // все игры должны быть проиграны
	moddedGame := MinimaxGame{
		Table: game.Table,
		State: 1,
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if moddedGame.Table[i][j] == 0 {
				moddedGame.Table[i][j] = 1

				endgame := CalculateEndgame(moddedGame)
				if endgame == 1 {
					//moddedGame.Table[i][j] = 0
					moddedGame.State = 0
					return moddedGame
				} else if endgame == -1 {
					moddedGame.State = 1
					return moddedGame
				}
				newGame := Minimax(moddedGame, 2)
				if newGame.State == 0 {
					//moddedGame.Table[i][j] = 0
					moddedGame.State = 0
					return moddedGame
				}

				moddedGame.Table[i][j] = 0
			}
		}
	}
	moddedGame.State = 1
	return moddedGame
}

// win: 1\2 ; have no space: -1 ; have space, noone win: 0
func CalculateEndgame(game MinimaxGame) int {
	for i := 0; i < 3; i++ {
		rowVal := game.Table[i][0]
		rowSame := true
		for j := 0; j < 3; j++ {
			if game.Table[i][j] != rowVal {
				rowSame = false
			}
		}

		if rowSame && rowVal != 0 {
			return int(rowVal)
		}
	}

	for i := 0; i < 3; i++ {
		columnVal := game.Table[0][i]
		columnSame := true
		for j := 0; j < 3; j++ {
			if game.Table[j][i] != columnVal {
				columnSame = false
			}
		}

		if columnSame && columnVal != 0 {
			return int(columnVal)
		}
	}

	midVal := game.Table[1][1]
	if game.Table[0][0] == midVal && game.Table[2][2] == midVal && midVal != 0 {
		return int(midVal)
	} else if game.Table[0][2] == midVal && game.Table[2][0] == midVal && midVal != 0 {
		return int(midVal)
	}

	haveSpace := false
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if game.Table[i][j] == 0 {
				haveSpace = true
				break
			}
		}
	}
	if haveSpace {
		return 0
	} else {
		return -1
	}
}
