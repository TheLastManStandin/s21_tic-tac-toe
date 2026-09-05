package gameRepository

import (
	"context"
	"tickTackToe/internal/domain"
)

type GameRepository interface {
	SaveGame(ctx context.Context, game domain.GameInstance, UUID string) error
	FindByID(ctx context.Context, UUID string) (domain.GameInstance, error)
}
