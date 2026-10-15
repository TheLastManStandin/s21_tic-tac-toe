package repositorys

import (
	"context"
	"tickTackToe/internal/domain"
)

type GameRepository interface {
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

	ConnectToGame(
		ctx context.Context,
		gameUUID string,
		playerUUID string,
	) error
}
