package AI

import "domain"

func Minimax(s domain.Service, game MinimaxGame, turn int) MinimaxGame {
	if turn == 2 { // хотя бы один путь
		moddedGame := MinimaxGame{
			Table: game.Table,
			State: 0,
		}

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if moddedGame.Table[i][j] == 0 {
					moddedGame.Table[i][j] = 2

					endgame := s.GameEnded(moddedGame.ToDomainGame())
					if endgame == 2 || endgame == -1 {
						//moddedGame.Table[i][j] = 0
						moddedGame.State = 1
						return moddedGame
					}
					newGame := Minimax(s, moddedGame, 1)
					if newGame.State == 1 {
						//moddedGame.Table[i][j] = 0
						moddedGame.State = 1
						return moddedGame
					}

					moddedGame.Table[i][j] = 0
				}
			}
		}
		moddedGame.State = 1
		return moddedGame
	} // все игры должны быть проиграны
	moddedGame := MinimaxGame{
		Table: game.Table,
		State: 1,
	}

	counter := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if moddedGame.Table[i][j] == 0 {
				counter++
				moddedGame.Table[i][j] = 1

				endgame := s.GameEnded(moddedGame.ToDomainGame())
				if endgame == 1 {
					//moddedGame.Table[i][j] = 0
					moddedGame.State = 0
					return moddedGame
				} else if endgame == -1 {
					moddedGame.State = 1
					return moddedGame
				}
				newGame := Minimax(s, moddedGame, 2)
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
