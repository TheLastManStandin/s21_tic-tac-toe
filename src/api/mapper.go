package api

import "domain"

func (gr GameRequest) ToDomainGame() domain.Game {
	res := domain.Game{}
	res.Table = domain.Table(gr.Board)
	res.UUID = gr.ID

	return res
}
