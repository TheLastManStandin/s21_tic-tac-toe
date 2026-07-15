package web

import "domain"

func (gr *GameRequest) toDomain() domain.GameRequest {
	return domain.GameRequest{
		X: gr.X,
		Y: gr.Y,
	}
}
