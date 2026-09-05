package datasource

import "tickTackToe/internal/domain"

func gameToRow(game domain.GameInstance) gameRow {
	table := make([][]int32, 3)

	for y := 0; y < 3; y++ {
		table[y] = make([]int32, 3)

		for x := 0; x < 3; x++ {
			table[y][x] = int32(game.Table[y][x])
		}
	}

	return gameRow{
		UUID:      game,
		GameTable: table,
		GameState: int(game.State),
	}
}

func rowToGame(row gameRow) domain.Game {
	game := domain.Game{
		ID:    row.UUID,
		State: domain.GameState(row.GameState),
	}

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			game.Table[y][x] = int(row.GameTable[y][x])
		}
	}

	return game
}
