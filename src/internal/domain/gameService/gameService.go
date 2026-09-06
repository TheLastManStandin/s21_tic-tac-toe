package gameService

import (
	"context"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/gameRepository"
	"tickTackToe/internal/domain/gameService/minimax"
)

type GameService struct {
	Repo gameRepository.GameRepository
}

func NewGameService(repo gameRepository.GameRepository) *GameService {
	return &GameService{
		Repo: repo,
	}
}

func (gs *GameService) GetGame(ctx context.Context, UUID string) (domain.GameInstance, error) {
	return gs.Repo.FindByID(ctx, UUID)
}

func (gs *GameService) MinimaxTurn(ctx context.Context, game domain.GameInstance) (domain.GameInstance, error) {
	internalGame := minimax.GameFromDomain(game)
	minimax.Minimax(&internalGame, 2)
	game = minimax.GameToDomain(internalGame, game)

	return game, gs.SaveGame(ctx, game)
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
