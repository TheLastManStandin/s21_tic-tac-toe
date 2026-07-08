package api

import (
	"domain"
)

type Handler struct {
	service domain.GameService
}

func NewHandler(service domain.GameService) *Handler {
	return &Handler{
		service: service,
	}
}
