package handlers

import (
	"encoding/json"
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/web/httpresponse"
	"tickTackToe/internal/web/requestcontext"

	"go.uber.org/zap"
)

func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	playerUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok || playerUUID == "" {
		httpresponse.WriteUnauthorized(w)
		return
	}
	games, err := h.service.GetGames(r.Context(), playerUUID)
	if err != nil {
		h.logger.Error("failed to list games", zap.Error(err))
		httpresponse.WriteError(w, err)
		return
	}
	response := make([]GameDetailsResponse, 0, len(games))
	for _, game := range games {
		response = append(response, gameDetailsResponse(game))
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to write games response", zap.Error(err))
	}
}

func gameDetailsResponse(game domain.GameInstance) GameDetailsResponse {
	return GameDetailsResponse{
		UUID: game.UUID, Table: game.Table, State: game.State.String(),
		Minimax: game.Minimax, FirstPlayerUUID: game.FirstPlayerUUID,
		SecondPlayerUUID: game.SecondPlayerUUID,
	}
}
