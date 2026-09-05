package gameService

import "tickTackToe/internal/domain/gameRepository"

type GameService struct {
	Repo gameRepository.GameRepository
}

func NewGameService(repo gameRepository.GameRepository) *GameService {
	return &GameService{
		Repo: repo,
	}
}
