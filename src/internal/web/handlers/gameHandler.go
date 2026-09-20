package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/requestcontext"

	"go.uber.org/zap"
)

type GameHandler struct {
	service *services.GameService
	logger  *zap.Logger
}

func NewGameHandler(service *services.GameService, logger *zap.Logger) *GameHandler {
	return &GameHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GameHandler) MakeTurn(w http.ResponseWriter, r *http.Request) {
	gameUUID := r.PathValue("UUID")
	defer r.Body.Close()
	h.logger.Info(fmt.Sprintf("Turning game %s", gameUUID))

	// TODO: Авторизация (может ли пользователь играть эту игру)

	request, err := decodeValidateRequest(r)
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.MakeTurn(r.Context(), gameUUID, request.X, request.Y)
	if err != nil {
		http.Error(w, "failed to make turn: "+err.Error(), http.StatusInternalServerError)
	}
	writeGameResponse(w, game)
}

func writeGameResponse(w http.ResponseWriter, game domain.GameInstance) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(GameResponse{
		Table: game.Table,
		State: game.State.String(),
	})
}

func decodeValidateRequest(r *http.Request) (GameRequest, error) {
	var req GameRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return GameRequest{}, fmt.Errorf("decode JSON: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return GameRequest{}, errors.New("body must contain exactly one JSON object")
	}

	if req.X < 0 || req.X > 2 {
		return GameRequest{}, errors.New("X must be between 0 and 2")
	}

	if req.Y < 0 || req.Y > 2 {
		return GameRequest{}, errors.New("Y must be between 0 and 2")
	}

	return req, nil
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	uuid, ok := requestcontext.UserUUID(r.Context())
	if !ok {
		h.logger.Error("failed to find user uuid", zap.String("uuid", uuid))
		http.Error(w, "no user uuid found", http.StatusUnauthorized)
	}
	game, err := h.service.InitGame(r.Context(), false, uuid)

	if err != nil {
		h.logger.Error("failed to initialize game", zap.Error(err))
		http.Error(
			w,
			"failed to initialize game: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	writeGameResponse(w, game)
}
