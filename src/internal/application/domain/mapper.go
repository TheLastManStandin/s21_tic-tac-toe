package domain

import "datasource"
import "AI"

func dbGameToGame(dbGame datasource.DbGame) Game {
	return Game{
		UUID:  dbGame.UUID,
		Table: dbGame.Table,
	}
}

func DomainToDbGame(dbGame Game) datasource.DbGame {
	return datasource.DbGame{
		UUID:  dbGame.UUID,
		Table: dbGame.Table,
	}
}

func GameToMinimaxGame(g Game) AI.MinimaxGame {
	return AI.MinimaxGame{
		Table: g.Table,
	}
}

func AIMinimaxGameToGame(g AI.MinimaxGame) Game {
	return Game{
		Table: g.Table,
	}
}
