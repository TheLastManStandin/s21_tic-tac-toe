package AI

import "domain"

type MinimaxGame struct {
	Table domain.Table
	State int // 1:good 0:bad
}

func (m MinimaxGame) ToDomainGame() domain.Game {
	out := domain.Game{
		Table: m.Table,
	}

	return out
}
