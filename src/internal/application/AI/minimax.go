package AI

import "domain"

func Minimax(game domain.Game, turn int8) domain.Game {
	res := domain.Game{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if turn == 1 && game.Table[i][j] == 0 {
				game.Table[i][j] = 1
				res = Minimax(game, 2)
			} else if turn == 2 && game.Table[i][j] == 0 {
				game.Table[i][j] = 2
				res = Minimax(game, 1)
			}
		}
	}
}
