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
