package handlers

import (
	"context"
	"errors"
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
	initGameCalls int
	gameUUID      string
	playerUUID    string
	minimax       bool
	initGameErr   error
}

func (r *gameRepositoryStub) SaveGame(context.Context, domain.GameInstance) error {
	return nil
}

func (r *gameRepositoryStub) FindGameByID(
	context.Context,
	string,
) (domain.GameInstance, error) {
	return domain.GameInstance{}, nil
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
