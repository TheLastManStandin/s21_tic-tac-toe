package api

import "domain"

func (gr GameRequest) ToDomainGame(UUID string) domain.Game {
	res := domain.Game{}
	res.Table = gr.Table
	res.UUID = UUID

	return res
}
