package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tickTackToe/internal/datasource"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/gameService"
	"tickTackToe/internal/web"

	"go.uber.org/zap"
)

type GameHandler struct {
	service *gameService.GameService
	logger  *zap.Logger
}

func NewGameHandler(service *gameService.GameService, logger *zap.Logger) *GameHandler {
	return &GameHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GameHandler) MakeTurn(w http.ResponseWriter, r *http.Request) {
	gameUUID := r.PathValue("UUID")
	defer r.Body.Close()
	ctx := context.WithValue(r.Context(), "logger", h.logger)
	h.logger.Info(fmt.Sprintf("Turning game %s", gameUUID))

	// TODO: Авторизация (может ли пользователь играть эту игру)

	request, err := decodeValidateRequest(r)
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверка статуса игры (Получение игры)
	domainGame, err := h.service.GetGame(ctx, gameUUID)
	if err != nil {
		var gameNotFound datasource.DatabaseGameNotFount
		if !errors.As(err, &gameNotFound) {
			http.Error(w, "get game: "+err.Error(), http.StatusInternalServerError)
			return
		}

		domainGame, err = h.service.InitGame(ctx, gameUUID)
		if err != nil {
			http.Error(w, "init game: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if domainGame.State == domain.FirstPlayerWon ||
		domainGame.State == domain.SecondPlayerWon ||
		domainGame.State == domain.Draw {
		writeGameResponse(w, domainGame)
		return
	}
	if domainGame.Table[request.X][request.Y] != 0 {
		http.Error(w, "cell is already occupied", http.StatusConflict)
		return
	}

	webGame := web.GameFromDomain(domainGame)

	// Модификация игры
	webGame.Table[request.X][request.Y] = 1 // TODO вынести в domain слой
	moddedDomainGame := web.GameToDomain(webGame)
	moddedDomainGame.State = h.service.CalculateGameState(moddedDomainGame, domain.SecondPlayerTurn)
	if moddedDomainGame.State == domain.FirstPlayerWon ||
		moddedDomainGame.State == domain.SecondPlayerWon ||
		moddedDomainGame.State == domain.Draw {
		writeGameResponse(w, moddedDomainGame)
		return
	}

	moddedDomainGame, err = h.service.MinimaxTurn(ctx, moddedDomainGame)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	moddedDomainGame.State = h.service.CalculateGameState(moddedDomainGame, domain.FirstPlayerTurn)
	writeGameResponse(w, moddedDomainGame)

	// Сохранение игры
	//h.service.SaveGame(r.Context(), moddedDomainGame)
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
