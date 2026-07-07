package domain

import "datasource"

func dbGameToGame(dbGame datasource.DbGame) Game {
	return Game{
		UUID:  dbGame.UUID,
		Table: dbGame.Table,
	}
}
