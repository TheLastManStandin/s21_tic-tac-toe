package handlers

import (
	"net/http"
	"tickTackToe/internal/domain/gameService"
)

type GameHandler struct {
	service *gameService.GameService
}

func NewGameHandler(service *gameService.GameService) *GameHandler {
	return &GameHandler{
		service: service,
	}
}

func (h *GameHandler) MakeTurn(w http.ResponseWriter, r *http.Request) {
	gameUUID := r.PathValue("uuid")

	w.Write([]byte(gameUUID))
}
