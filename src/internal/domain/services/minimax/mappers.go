package minimax

import "tickTackToe/internal/domain"

func GameFromDomain(game domain.GameInstance) minimaxGame {
	return minimaxGame{
		table:      game.Table,
		worstMarch: Still,
	}
}

func GameToDomain(game minimaxGame, source domain.GameInstance) domain.GameInstance {
	source.Table = game.table
	return source
}
