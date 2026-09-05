package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tickTackToe/internal/domain/gameService"
	"tickTackToe/internal/web"
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
	defer r.Body.Close()

	// TODO: Авторизация (может ли пользователь играть эту игру)

	request, err := decodeValidateRequest(r)
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверка статуса игры (Получение игры)
	domainGame, err := h.service.GetGame(r.Context(), gameUUID)
	if err != nil {
		http.Error(w, "get game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	webGame := web.GameFromDomain(domainGame)

	// Модификация игры
	_ = request
	_ = webGame

	// Сохранение игры
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
