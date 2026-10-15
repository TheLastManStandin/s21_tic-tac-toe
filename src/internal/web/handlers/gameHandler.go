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

	"github.com/google/uuid"
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
	userUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	h.logger.Info(fmt.Sprintf("Turning game %s", gameUUID))

	request, err := decodeValidateTurnRequest(r)
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.MakeTurn(r.Context(), gameUUID, userUUID, request.X, request.Y)
	if err != nil {
		status := http.StatusInternalServerError
		message := "internal server error"

		switch {
		case errors.Is(err, domain.ErrGameAccessDenied):
			status = http.StatusForbidden
			message = err.Error()

		case errors.Is(err, domain.ErrGameNotFound):
			status = http.StatusNotFound
			message = err.Error()

		case errors.Is(err, domain.ErrNotYourTurn),
			errors.Is(err, domain.ErrWaitingForPlayers),
			errors.Is(err, domain.ErrCellOccupied):
			status = http.StatusConflict
			message = err.Error()
		}

		h.logger.Error("failed to make turn", zap.Error(err))
		http.Error(w, message, status)
		return

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

func decodeValidateTurnRequest(r *http.Request) (TurnRequest, error) {
	var req TurnRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return TurnRequest{}, fmt.Errorf("decode JSON: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return TurnRequest{}, errors.New("body must contain exactly one JSON object")
	}

	if req.X < 0 || req.X > 2 {
		return TurnRequest{}, errors.New("X must be between 0 and 2")
	}

	if req.Y < 0 || req.Y > 2 {
		return TurnRequest{}, errors.New("Y must be between 0 and 2")
	}

	return req, nil
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	uuid, ok := requestcontext.UserUUID(r.Context())
	if !ok {
		h.logger.Error("failed to find user uuid", zap.String("uuid", uuid))
		http.Error(w, "no user uuid found", http.StatusUnauthorized)
		return
	}

	request, err := decodeValidateRobotRequest(r)
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.InitGame(r.Context(), request.WithRobot, uuid)
	if err != nil {
		h.logger.Error("failed to initialize game", zap.Error(err))
		http.Error(
			w,
			"failed to initialize game: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	_, _ = w.Write([]byte(game.UUID + "\n"))
}

func decodeValidateRobotRequest(r *http.Request) (RobotRequest, error) {
	var body struct {
		WithRobot *bool `json:"WithRobot"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&body); err != nil {
		return RobotRequest{}, fmt.Errorf("decode JSON: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return RobotRequest{}, errors.New("body must contain exactly one JSON object")
	}

	if body.WithRobot == nil {
		return RobotRequest{}, errors.New("WithRobot is required and must be a boolean")
	}

	return RobotRequest{WithRobot: *body.WithRobot}, nil
}

func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	playerUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok || playerUUID == "" {
		h.logger.Error("failed to find user uuid", zap.String("uuid", playerUUID))
		http.Error(w, "no user uuid found", http.StatusUnauthorized)
		return
	}

	gameUUID := r.PathValue("UUID")

	if _, err := uuid.Parse(gameUUID); err != nil {
		http.Error(w, "invalid game UUID", http.StatusBadRequest)
		return
	}

	if err := h.service.JoinGame(r.Context(), gameUUID, playerUUID); err != nil {
		if errors.Is(err, domain.ErrCannotConnectToGame) {
			http.Error(w, domain.ErrCannotConnectToGame.Error(), http.StatusConflict)
			return
		}

		h.logger.Error("failed to join game", zap.String("game_uuid", gameUUID), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
