package domain

import (
	"datasource"
	"errors"
)

func HandleTurn(req GameRequest, db datasource.Database, UUID string) (GameInstance, error) {
	loadedGame, err := db.LoadGame(UUID)
	var savedGame GameInstance
	if err != nil {
		savedGame = gameInit(req, UUID)
	} else {
		savedGame = gameFromDatasourceToDomain(loadedGame, UUID)

		if savedGame.Table[req.Y][req.X] == 0 {
			savedGame.Table[req.Y][req.X] = 1
		} else {
			savedGame.Status = calculateEndgame(savedGame.toMinimax())
			if savedGame.Status != Still {
				return savedGame, nil
			}
			return GameInstance{}, errors.New("invalid turn")
		}
	}

	if status := calculateEndgame(savedGame.toMinimax()); status >= 0 {
		savedGame.Status = status
		err = db.SaveGame(savedGame.toDatasource(), UUID)
		if err != nil {
			return GameInstance{}, err
		}
		return savedGame, nil
	}

	nextTurn := savedGame.toMinimax()
	minimax(&nextTurn, 2)
	savedGame = nextTurn.toDomain(UUID)
	err = db.SaveGame(savedGame.toDatasource(), UUID)
	if err != nil {
		return GameInstance{}, err
	}

	return savedGame, nil
}

func gameInit(req GameRequest, UUID string) GameInstance {
	table := [3][3]int{}
	table[req.Y][req.X] = 1
	return GameInstance{
		Table:  table,
		UUID:   UUID,
		Status: Still,
	}
}

// Возвращает 1 или 2, если победил игрок(1) или минимакс(2).
// Возвращает 0 в случаи ничьи.
// Возвращает -1 если никто не победил и игра продолжается
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
