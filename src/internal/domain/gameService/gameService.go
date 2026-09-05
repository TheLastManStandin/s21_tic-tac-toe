package gameService

import (
	"context"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/gameRepository"
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
