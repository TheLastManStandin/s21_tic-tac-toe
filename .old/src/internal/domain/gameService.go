package domain

import (
	"datasource"
	"errors"
)

type GameService struct {
	Database datasource.Database
}

func NewGameService(db datasource.Database) *GameService {
	return &GameService{
		Database: db,
	}
}

func (gs GameService) GetUpdatedGame(gameUUID string, turnRequest GameRequest) (GameInstance, error) {
	loadedGame, err := gs.Database.LoadGame(gameUUID)
	var savedGame GameInstance
	if err != nil {
		savedGame = gameInit(turnRequest, gameUUID)
	} else {
		savedGame = gameFromDatasourceToDomain(loadedGame, gameUUID)

		if savedGame.State >= datasource.FirstPlayerWon {
			return savedGame, nil
		}

		if savedGame.Table[turnRequest.Y][turnRequest.X] == 0 {
			savedGame.Table[turnRequest.Y][turnRequest.X] = 1
		} else {
			return GameInstance{}, errors.New("invalid turn")
		}
	}

	// todo придумать как получать ход (сейчас 2 хардкод)
	if status := gs.GetGameState(savedGame, 2); status >= datasource.FirstPlayerWon {
		savedGame.State = status
		err = gs.Database.SaveGame(savedGame.toDatasource(), gameUUID)
		if err != nil {
			return GameInstance{}, err
		}
		return savedGame, nil
	}

	return savedGame, nil
}

func (gs GameService) GetGameState(game GameInstance, turn int) datasource.GameState {
	for i := 0; i < 3; i++ {
		rowVal := game.Table[i][0]
		rowSame := true
		for j := 0; j < 3; j++ {
			if game.Table[i][j] != rowVal {
				rowSame = false
			}
		}

		if rowSame && rowVal != 0 {
			if rowVal == 1 {
				return datasource.FirstPlayerWon
			} else if rowVal == 2 {
				return datasource.SecondPlayerWon
			}
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
			if columnVal == 1 {
				return datasource.FirstPlayerWon
			} else if columnVal == 2 {
				return datasource.SecondPlayerWon
			}
		}
	}

	midVal := game.Table[1][1]
	if game.Table[0][0] == midVal && game.Table[2][2] == midVal && midVal != 0 {
		if midVal == 1 {
			return datasource.FirstPlayerWon
		} else if midVal == 2 {
			return datasource.SecondPlayerWon
		}
	} else if game.Table[0][2] == midVal && game.Table[2][0] == midVal && midVal != 0 {
		if midVal == 1 {
			return datasource.FirstPlayerWon
		} else if midVal == 2 {
			return datasource.SecondPlayerWon
		}
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
		if turn == 1 {
			return datasource.FirstPlayerTurn
		} else {
			return datasource.SecondPlayerTurn
		}
	} else {
		return datasource.Draw
	}
}

func (gs GameService) MinimaxTurn(game GameInstance) error {
	// todo переписать
	nextTurn := game.toMinimax()
	minimax(&nextTurn, 2)
	game = nextTurn.toDomain(game.UUID, 0)
	game.State = gs.GetGameState(game, 1)
	err := gs.Database.SaveGame(game.toDatasource(), game.UUID)
	if err != nil {
		return err
	}
	return nil
}

//MinimaxTurn
//Новые функции из 7 задания
