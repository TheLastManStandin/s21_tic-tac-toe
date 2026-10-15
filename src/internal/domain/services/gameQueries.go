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

func (gs *GameService) GetGameStatus(ctx context.Context, gameUUID, playerUUID string) (domain.GameInstance, error) {
	if playerUUID == "" {
		return domain.GameInstance{}, domain.ErrGameAccessDenied
	}
	game, err := gs.GetGame(ctx, gameUUID)
	if err != nil {
		return domain.GameInstance{}, fmt.Errorf("get game status: %w", err)
	}
	if playerUUID != game.FirstPlayerUUID && playerUUID != game.SecondPlayerUUID {
		return domain.GameInstance{}, domain.ErrGameAccessDenied
	}
	return game, nil
}
