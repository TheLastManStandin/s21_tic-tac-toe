package repositorys

import (
	"context"
	"tickTackToe/internal/domain"
)

type GameRepository interface {
	ListAvailableGames(ctx context.Context, playerUUID string) ([]domain.GameInstance, error)

	SaveGame(
		ctx context.Context,
		game domain.GameInstance,
	) error

	FindGameByID(
		ctx context.Context,
		UUID string,
	) (domain.GameInstance, error)

	InitGame(
		ctx context.Context,
		gameUUID string,
		playerUUID string,
		minimax bool,
	) error

	JoinGame(
		ctx context.Context,
		gameUUID string,
		playerUUID string,
	) error
}
