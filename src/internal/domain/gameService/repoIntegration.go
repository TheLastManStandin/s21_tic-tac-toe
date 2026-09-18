package gameService

import (
	"context"
	"tickTackToe/internal/domain"
)

func (gs *GameService) GetGame(ctx context.Context, UUID string) (domain.GameInstance, error) {
	return gs.Repo.FindByID(ctx, UUID)
}

func (gs *GameService) SaveGame(ctx context.Context, game domain.GameInstance) error {
	return gs.Repo.SaveGame(ctx, game)
}

func (gs *GameService) InitGame(ctx context.Context, UUID string) (domain.GameInstance, error) {
	game := domain.GameInstance{
		UUID:    UUID,
		State:   domain.WaitingForPlayers,
		Minimax: true,
	}

	if err := gs.Repo.InitGame(ctx, UUID); err != nil {
		return domain.GameInstance{}, err
	}

	return game, nil
}
