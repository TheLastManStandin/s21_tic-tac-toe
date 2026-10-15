package services

import (
	"context"
	"errors"
	"fmt"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/repositorys"
	minimax2 "tickTackToe/internal/domain/services/minimax"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type GameService struct {
	Repo   repositorys.GameRepository
	logger *zap.Logger
}

func NewGameService(repo repositorys.GameRepository, logger *zap.Logger) *GameService {
	return &GameService{
		Repo:   repo,
		logger: logger.Named("game_service"),
	}
}

func (gs *GameService) MinimaxTurn(ctx context.Context, game domain.GameInstance) domain.GameInstance {
	internalGame := minimax2.GameFromDomain(game)
	minimax2.Minimax(&internalGame, 2)
	game = minimax2.GameToDomain(internalGame, game)

	return game
}

func (gs *GameService) MakeTurn(
	ctx context.Context,
	gameUUID string,
	userUUID string,
	x, y int,
) (domain.GameInstance, error) {
	logger := gs.logger.With(
		zap.String("game_uuid", gameUUID),
		zap.Int("x", x),
		zap.Int("y", y),
	)
	logger.Info("making turn")

	// Получение игры
	domainGame, err := gs.GetGame(ctx, gameUUID)
	if err != nil {
		logger.Error("failed to get game", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GameInstance{}, domain.ErrGameNotFound
		}
		return domain.GameInstance{}, fmt.Errorf("failed to get game:%w", err)
	}

	// Узнаем можно ли пользователю делать ходы в этой игре и его номер(p1 / p2)
	var playerNumber int = 0
	if domainGame.FirstPlayerUUID == userUUID {
		playerNumber = 1
	} else if domainGame.SecondPlayerUUID == userUUID {
		playerNumber = 2
	}
	if playerNumber == 0 {
		return domain.GameInstance{}, domain.ErrGameAccessDenied
	}

	// Проверка статуса игры
	gameState := domainGame.State
	if gameState.IsFinished() {
		return domainGame, nil
	}

	// Проверка если клетка занята
	if domainGame.Table[x][y] != 0 {
		return domainGame, domain.ErrCellOccupied
	}

	// Модификация игры
	if domainGame.Minimax {
		domainGame.Table[x][y] = 1
		domainGame.State = gs.CalculateGameState(domainGame, domain.SecondPlayerTurn)
		if !domainGame.State.IsFinished() {
			domainGame = gs.MinimaxTurn(ctx, domainGame)
			domainGame.State = gs.CalculateGameState(domainGame, domain.FirstPlayerTurn)
			logger.Debug("minimax turn applied", zap.String("state", domainGame.State.String()))
		}
	} else {
		if gameState == domain.WaitingForPlayers {
			return domain.GameInstance{}, domain.ErrWaitingForPlayers
		}
		if gameState == domain.FirstPlayerTurn && playerNumber != 1 {
			return domain.GameInstance{}, domain.ErrNotYourTurn
		} else if gameState == domain.SecondPlayerTurn && playerNumber != 2 {
			return domain.GameInstance{}, domain.ErrNotYourTurn
		}

		domainGame.Table[x][y] = playerNumber
		domainGame.State = gs.CalculateGameState(domainGame, domain.SecondPlayerTurn)
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

func (gs *GameService) GetGame(ctx context.Context, UUID string) (domain.GameInstance, error) {
	return gs.Repo.FindGameByID(ctx, UUID)
}

func (gs *GameService) SaveGame(ctx context.Context, game domain.GameInstance) error {
	return gs.Repo.SaveGame(ctx, game)
}

func (gs *GameService) InitGame(
	ctx context.Context,
	minimax bool,
	playerUUID string,
) (domain.GameInstance, error) {
	game := domain.GameInstance{
		UUID:    uuid.New().String(),
		State:   domain.WaitingForPlayers,
		Minimax: minimax,
	}

	if err := gs.Repo.InitGame(ctx, game.UUID, playerUUID, minimax); err != nil {
		return domain.GameInstance{}, err
	}
	return game, nil
}

func (gs *GameService) JoinGame(ctx context.Context, gameUUID string, playerUUID string) error {
	err := gs.Repo.JoinGame(ctx, gameUUID, playerUUID)
	if err == nil {
		// return fmt.Errorf("join game: %w", err)
		return nil
	}
	if !errors.Is(err, domain.ErrCannotConnectToGame) {
		return fmt.Errorf("join game: %w", err)
	}

	game, err := gs.GetGame(ctx, gameUUID)
	if err != nil {
		return err
	}

	if game.Minimax {
		return domain.ErrCantConnectGameWithRobot
	}
	if game.SecondPlayerUUID == playerUUID || game.FirstPlayerUUID == playerUUID {
		return domain.ErrAlreadyInGame
	}
	if game.State != domain.WaitingForPlayers {
		return domain.ErrGameAlreadyStarted
	}

	return domain.ErrCannotConnectToGame
}
