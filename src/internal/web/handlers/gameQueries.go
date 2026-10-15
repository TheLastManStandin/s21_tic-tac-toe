package handlers

import (
	"encoding/json"
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/web/httpresponse"
	"tickTackToe/internal/web/requestcontext"

	"github.com/google/uuid"
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

func (h *GameHandler) GetGameStatus(w http.ResponseWriter, r *http.Request) {
	playerUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok || playerUUID == "" {
		httpresponse.WriteUnauthorized(w)
		return
	}
	gameUUID, err := uuid.Parse(r.PathValue("UUID"))
	if err != nil {
		http.Error(w, "invalid game UUID", http.StatusBadRequest)
		return
	}
	game, err := h.service.GetGameStatus(r.Context(), gameUUID.String(), playerUUID)
	if err != nil {
		h.logger.Info("failed to get game status", zap.Error(err))
		httpresponse.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gameDetailsResponse(game)); err != nil {
		h.logger.Error("failed to write game status", zap.Error(err))
	}
}

func gameDetailsResponse(game domain.GameInstance) GameDetailsResponse {
	return GameDetailsResponse{
		UUID: game.UUID, Table: game.Table, State: game.State.String(),
		Minimax: game.Minimax, FirstPlayerUUID: game.FirstPlayerUUID,
		SecondPlayerUUID: game.SecondPlayerUUID,
	}
}
