package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/requestcontext"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type gameRepositoryStub struct {
	initGameCalls  int
	gameUUID       string
	playerUUID     string
	minimax        bool
	initGameErr    error
	connectCalls   int
	connectErr     error
	foundGame      domain.GameInstance
	findGameErr    error
	availableGames []domain.GameInstance
	listGamesErr   error
	listGamesCalls int
}

func (r *gameRepositoryStub) ListAvailableGames(_ context.Context, playerUUID string) ([]domain.GameInstance, error) {
	r.listGamesCalls++
	r.playerUUID = playerUUID
	return r.availableGames, r.listGamesErr
}

func (r *gameRepositoryStub) JoinGame(_ context.Context, gameUUID, playerUUID string) error {
	r.connectCalls++
	r.gameUUID = gameUUID
	r.playerUUID = playerUUID
	return r.connectErr
}

func (r *gameRepositoryStub) SaveGame(context.Context, domain.GameInstance) error {
	return nil
}

func (r *gameRepositoryStub) FindGameByID(
	context.Context,
	string,
) (domain.GameInstance, error) {
	return r.foundGame, r.findGameErr
}

func (r *gameRepositoryStub) InitGame(
	_ context.Context,
	gameUUID string,
	playerUUID string,
	minimax bool,
) error {
	r.initGameCalls++
	r.gameUUID = gameUUID
	r.playerUUID = playerUUID
	r.minimax = minimax
	return r.initGameErr
}

func TestGameHandlerJoinGame(t *testing.T) {
	const gameUUID = "11111111-1111-4111-8111-111111111111"
	const playerUUID = "22222222-2222-4222-8222-222222222222"
	tests := []struct {
		name       string
		gameUUID   string
		playerUUID string
		err        error
		foundGame  domain.GameInstance
		findErr    error
		wantStatus int
		wantCalls  int
		wantBody   string
	}{
		{name: "success", gameUUID: gameUUID, playerUUID: playerUUID, wantStatus: http.StatusNoContent, wantCalls: 1},
		{name: "unauthenticated", gameUUID: gameUUID, wantStatus: http.StatusUnauthorized, wantBody: "unauthorized\n"},
		{name: "invalid UUID", gameUUID: "invalid", playerUUID: playerUUID, wantStatus: http.StatusBadRequest, wantBody: "invalid game UUID\n"},
		{name: "cannot join", gameUUID: gameUUID, playerUUID: playerUUID, err: fmt.Errorf("repository: %w", domain.ErrCannotConnectToGame), wantStatus: http.StatusConflict, wantCalls: 1, wantBody: "cannot connect to game\n"},
		{name: "game not found", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, findErr: fmt.Errorf("repository: %w", domain.ErrGameNotFound), wantStatus: http.StatusNotFound, wantCalls: 1, wantBody: domain.ErrGameNotFound.Error() + "\n"},
		{name: "game against robot", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, foundGame: domain.GameInstance{Minimax: true}, wantStatus: http.StatusConflict, wantCalls: 1, wantBody: domain.ErrCantConnectGameWithRobot.Error() + "\n"},
		{name: "game already started", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, foundGame: domain.GameInstance{State: domain.FirstPlayerTurn}, wantStatus: http.StatusConflict, wantCalls: 1, wantBody: domain.ErrGameAlreadyStarted.Error() + "\n"},
		{name: "already first player", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, foundGame: domain.GameInstance{FirstPlayerUUID: playerUUID}, wantStatus: http.StatusConflict, wantCalls: 1, wantBody: domain.ErrAlreadyInGame.Error() + "\n"},
		{name: "already second player", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, foundGame: domain.GameInstance{SecondPlayerUUID: playerUUID}, wantStatus: http.StatusConflict, wantCalls: 1, wantBody: domain.ErrAlreadyInGame.Error() + "\n"},
		{name: "lookup failure", gameUUID: gameUUID, playerUUID: playerUUID, err: domain.ErrCannotConnectToGame, findErr: errors.New("private database details"), wantStatus: http.StatusInternalServerError, wantCalls: 1, wantBody: "internal server error\n"},
		{name: "database error", gameUUID: gameUUID, playerUUID: playerUUID, err: errors.New("private database details"), wantStatus: http.StatusInternalServerError, wantCalls: 1, wantBody: "internal server error\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &gameRepositoryStub{connectErr: tt.err, foundGame: tt.foundGame, findGameErr: tt.findErr}
			service := services.NewGameService(repository, zap.NewNop())
			handler := NewGameHandler(service, zap.NewNop())
			request := httptest.NewRequest(http.MethodPost, "/game/"+tt.gameUUID+"/join", nil)
			if tt.playerUUID != "" {
				request = request.WithContext(requestcontext.WithUserUUID(request.Context(), tt.playerUUID))
			}
			response := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.HandleFunc("POST /game/{UUID}/join", handler.JoinGame)
			mux.ServeHTTP(response, request)
			if response.Code != tt.wantStatus || response.Body.String() != tt.wantBody {
				t.Fatalf("response = (%d, %q), want (%d, %q)", response.Code, response.Body.String(), tt.wantStatus, tt.wantBody)
			}
			if repository.connectCalls != tt.wantCalls {
				t.Fatalf("ConnectToGame calls = %d, want %d", repository.connectCalls, tt.wantCalls)
			}
			if tt.wantCalls > 0 && (repository.gameUUID != tt.gameUUID || repository.playerUUID != tt.playerUUID) {
				t.Fatalf("ConnectToGame received wrong identifiers: %q, %q", repository.gameUUID, repository.playerUUID)
			}
		})
	}
}

func TestGameHandlerCreateGame(t *testing.T) {
	t.Run("creates multiplayer game for authenticated user", func(t *testing.T) {
		const playerUUID = "11111111-1111-4111-8111-111111111111"

		repository := &gameRepositoryStub{}
		service := services.NewGameService(repository, zap.NewNop())
		handler := NewGameHandler(service, zap.NewNop())
		request := httptest.NewRequest(
			http.MethodPost,
			"/game",
			strings.NewReader(`{"WithRobot":false}`),
		)
		request = request.WithContext(
			requestcontext.WithUserUUID(request.Context(), playerUUID),
		)
		response := httptest.NewRecorder()

		handler.CreateGame(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if repository.initGameCalls != 1 {
			t.Fatalf("InitGame calls = %d, want 1", repository.initGameCalls)
		}
		if repository.playerUUID != playerUUID {
			t.Fatalf("player UUID = %q, want %q", repository.playerUUID, playerUUID)
		}
		if repository.minimax {
			t.Fatal("minimax = true, want false for multiplayer game")
		}

		responseUUID := strings.TrimSpace(response.Body.String())
		if _, err := uuid.Parse(responseUUID); err != nil {
			t.Fatalf("response UUID %q is invalid: %v", responseUUID, err)
		}
		if repository.gameUUID != responseUUID {
			t.Fatalf(
				"stored game UUID = %q, response UUID = %q",
				repository.gameUUID,
				responseUUID,
			)
		}
	})

	t.Run("rejects request without authenticated user", func(t *testing.T) {
		repository := &gameRepositoryStub{}
		service := services.NewGameService(repository, zap.NewNop())
		handler := NewGameHandler(service, zap.NewNop())
		request := httptest.NewRequest(
			http.MethodPost,
			"/game",
			strings.NewReader(`{"WithRobot":false}`),
		)
		response := httptest.NewRecorder()

		handler.CreateGame(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
		if repository.initGameCalls != 0 {
			t.Fatalf(
				"InitGame calls = %d, want 0 for unauthenticated request",
				repository.initGameCalls,
			)
		}
	})

	t.Run("returns server error when game cannot be stored", func(t *testing.T) {
		const playerUUID = "22222222-2222-4222-8222-222222222222"

		repository := &gameRepositoryStub{initGameErr: errors.New("database unavailable")}
		service := services.NewGameService(repository, zap.NewNop())
		handler := NewGameHandler(service, zap.NewNop())
		request := httptest.NewRequest(
			http.MethodPost,
			"/game",
			strings.NewReader(`{"WithRobot":false}`),
		)
		request = request.WithContext(
			requestcontext.WithUserUUID(request.Context(), playerUUID),
		)
		response := httptest.NewRecorder()

		handler.CreateGame(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf(
				"status = %d, want %d",
				response.Code,
				http.StatusInternalServerError,
			)
		}
		if repository.initGameCalls != 1 {
			t.Fatalf("InitGame calls = %d, want 1", repository.initGameCalls)
		}
	})

	t.Run("creates game against robot", func(t *testing.T) {
		const playerUUID = "33333333-3333-4333-8333-333333333333"

		repository := &gameRepositoryStub{}
		service := services.NewGameService(repository, zap.NewNop())
		handler := NewGameHandler(service, zap.NewNop())
		request := httptest.NewRequest(
			http.MethodPost,
			"/game",
			strings.NewReader(`{"WithRobot":true}`),
		)
		request = request.WithContext(
			requestcontext.WithUserUUID(request.Context(), playerUUID),
		)
		response := httptest.NewRecorder()

		handler.CreateGame(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if !repository.minimax {
			t.Fatal("minimax = false, want true for game against robot")
		}
	})

	t.Run("rejects invalid robot option", func(t *testing.T) {
		const playerUUID = "44444444-4444-4444-8444-444444444444"

		repository := &gameRepositoryStub{}
		service := services.NewGameService(repository, zap.NewNop())
		handler := NewGameHandler(service, zap.NewNop())
		request := httptest.NewRequest(
			http.MethodPost,
			"/game",
			strings.NewReader(`{"WithRobot":"true"}`),
		)
		request = request.WithContext(
			requestcontext.WithUserUUID(request.Context(), playerUUID),
		)
		response := httptest.NewRecorder()

		handler.CreateGame(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
		}
		if repository.initGameCalls != 0 {
			t.Fatalf(
				"InitGame calls = %d, want 0 for invalid request",
				repository.initGameCalls,
			)
		}
	})
}

func TestDecodeValidateRobotRequest(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantRobot bool
		wantError bool
	}{
		{name: "robot enabled", body: `{"WithRobot":true}`, wantRobot: true},
		{name: "robot disabled", body: `{"WithRobot":false}`},
		{name: "missing field", body: `{}`, wantError: true},
		{name: "null field", body: `{"WithRobot":null}`, wantError: true},
		{name: "string instead of bool", body: `{"WithRobot":"true"}`, wantError: true},
		{name: "unknown field", body: `{"WithRobot":true,"Other":false}`, wantError: true},
		{name: "multiple objects", body: `{"WithRobot":true}{"WithRobot":false}`, wantError: true},
		{name: "empty body", body: ``, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/game",
				strings.NewReader(tt.body),
			)

			got, err := decodeValidateRobotRequest(request)
			if tt.wantError {
				if err == nil {
					t.Fatalf("decodeValidateRobotRequest() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeValidateRobotRequest() error = %v", err)
			}
			if got.WithRobot != tt.wantRobot {
				t.Fatalf("WithRobot = %t, want %t", got.WithRobot, tt.wantRobot)
			}
		})
	}
}
