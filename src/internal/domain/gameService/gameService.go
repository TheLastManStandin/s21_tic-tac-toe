package gameService

import (
	"context"
	"errors"
	"tickTackToe/internal/datasource"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/gameRepository"
	"tickTackToe/internal/domain/gameService/minimax"

	"go.uber.org/zap"
)

type GameService struct {
	Repo   gameRepository.GameRepository
	logger *zap.Logger
}

func NewGameService(repo gameRepository.GameRepository, logger *zap.Logger) *GameService {
	return &GameService{
		Repo:   repo,
		logger: logger.Named("game_service"),
	}
}

func (gs *GameService) MinimaxTurn(ctx context.Context, game domain.GameInstance) domain.GameInstance {
	internalGame := minimax.GameFromDomain(game)
	minimax.Minimax(&internalGame, 2)
	game = minimax.GameToDomain(internalGame, game)

	return game
}

func (gs *GameService) MakeTurn(
	ctx context.Context,
	gameUUID string,
	x, y int,
) (domain.GameInstance, error) {
	logger := gs.logger.With(
		zap.String("game_uuid", gameUUID),
		zap.Int("x", x),
		zap.Int("y", y),
	)
	logger.Info("making turn")

	// Проверка статуса игры (Получение игры)
	domainGame, err := gs.GetGame(ctx, gameUUID)
	if err != nil {
		var gameNotFound datasource.DatabaseGameNotFount
		if !errors.As(err, &gameNotFound) {
			logger.Error("failed to get game", zap.Error(err))
			return domain.GameInstance{}, errors.New("get game: " + err.Error())
		}

		logger.Info("game not found, initializing")
		domainGame, err = gs.InitGame(ctx, gameUUID)
		if err != nil {
			logger.Error("failed to initialize game", zap.Error(err))
			return domain.GameInstance{}, errors.New("init game: " + err.Error())
		}
	}
	if domainGame.State.IsFinished() {
		return domainGame, nil
	}
	if domainGame.Table[x][y] != 0 {
		return domain.GameInstance{}, errors.New("cell is already occupied")
	}

	// Модификация игры
	domainGame.Table[x][y] = 1
	domainGame.State = gs.CalculateGameState(domainGame, domain.SecondPlayerTurn)
	if !domainGame.State.IsFinished() {
		domainGame = gs.MinimaxTurn(ctx, domainGame)
		domainGame.State = gs.CalculateGameState(domainGame, domain.FirstPlayerTurn)
		logger.Debug("minimax turn applied", zap.String("state", domainGame.State.String()))
	}

	// Сохранение игры
	err = gs.SaveGame(ctx, domainGame)
	if err != nil {
		logger.Error("failed to save game", zap.Error(err))
		return domain.GameInstance{}, errors.New("database error: " + err.Error())
	}
	logger.Info("turn completed", zap.String("state", domainGame.State.String()))

	return domainGame, nil
}

func (gs *GameService) CalculateGameState(
	game domain.GameInstance,
	nextTurn domain.GameState,
) domain.GameState {
	for row := 0; row < 3; row++ {
		if winner := lineWinner(
			game.Table[row][0],
			game.Table[row][1],
			game.Table[row][2],
		); winner != 0 {
			return winnerState(winner)
		}
	}

	for column := 0; column < 3; column++ {
		if winner := lineWinner(
			game.Table[0][column],
			game.Table[1][column],
			game.Table[2][column],
		); winner != 0 {
			return winnerState(winner)
		}
	}

	if winner := lineWinner(
		game.Table[0][0],
		game.Table[1][1],
		game.Table[2][2],
	); winner != 0 {
		return winnerState(winner)
	}

	if winner := lineWinner(
		game.Table[0][2],
		game.Table[1][1],
		game.Table[2][0],
	); winner != 0 {
		return winnerState(winner)
	}

	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			if game.Table[row][column] == 0 {
				return nextTurn
			}
		}
	}

	return domain.Draw
}

func lineWinner(first, second, third int) int {
	if first != 0 && first == second && second == third {
		return first
	}

	return 0
}

func winnerState(winner int) domain.GameState {
	if winner == 1 {
		return domain.FirstPlayerWon
	}

	return domain.SecondPlayerWon
}
