package services

import (
	"context"
	"fmt"
	"tickTackToe/internal/domain"
)

func (gs *GameService) GetGames(ctx context.Context, playerUUID string) ([]domain.GameInstance, error) {
	if playerUUID == "" {
		return nil, domain.ErrGameAccessDenied
	}
	games, err := gs.Repo.ListAvailableGames(ctx, playerUUID)
	if err != nil {
		return nil, fmt.Errorf("list available games: %w", err)
	}
	return games, nil
}
