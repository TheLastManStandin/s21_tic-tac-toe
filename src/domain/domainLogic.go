package domain

import (
	"datasource"
)

func gameInit(req GameRequest, UUID string) GameInstance {
	table := [3][3]int{}
	table[req.Y][req.X] = 1
	return GameInstance{
		Table: table,
		UUID:  UUID,
		State: datasource.WaitingForPlayers,
	}
}

// Возвращает 1 или 2, если победил игрок(1) или минимакс(2).
// Возвращает 0 в случаи ничьи.
// Возвращает -1 если никто не победил и игра продолжается
//func calculateEndgame(game minimaxGame) MinimaxStatus {
//
//}
