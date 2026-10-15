package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/requestcontext"

	"go.uber.org/zap"
)

func TestGameHandlerGetGameStatus(t *testing.T) {
	const gameUUID = "11111111-1111-4111-8111-111111111111"
	const first = "22222222-2222-4222-8222-222222222222"
	const second = "33333333-3333-4333-8333-333333333333"
	for _, tt := range []struct {
		name   string
		id     string
		player string
		state  domain.GameState
		err    error
		status int
		calls  int
	}{
		{name: "first player", id: gameUUID, player: first, state: domain.SecondPlayerTurn, status: http.StatusOK, calls: 1},
		{name: "second player", id: gameUUID, player: second, state: domain.FirstPlayerTurn, status: http.StatusOK, calls: 1},
		{name: "finished game", id: gameUUID, player: first, state: domain.Draw, status: http.StatusOK, calls: 1},
		{name: "outsider", id: gameUUID, player: "44444444-4444-4444-8444-444444444444", status: http.StatusForbidden, calls: 1},
		{name: "missing game", id: gameUUID, player: first, err: fmt.Errorf("repository: %w", domain.ErrGameNotFound), status: http.StatusNotFound, calls: 1},
		{name: "database failure", id: gameUUID, player: first, err: errors.New("private database details"), status: http.StatusInternalServerError, calls: 1},
		{name: "invalid UUID", id: "invalid", player: first, status: http.StatusBadRequest},
		{name: "unauthenticated", id: gameUUID, status: http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			game := domain.GameInstance{UUID: gameUUID, FirstPlayerUUID: first, SecondPlayerUUID: second, State: tt.state, Table: [3][3]int{{1, 2, 0}}}
			repo := &gameRepositoryStub{foundGame: game, findGameErr: tt.err}
			handler := NewGameHandler(services.NewGameService(repo, zap.NewNop()), zap.NewNop())
			request := httptest.NewRequest(http.MethodGet, "/game/"+tt.id, nil)
			if tt.player != "" {
				request = request.WithContext(requestcontext.WithUserUUID(request.Context(), tt.player))
			}
			response := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.HandleFunc("GET /game/{UUID}", handler.GetGameStatus)
			mux.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body.String())
			}
			if repo.findGameCalls != tt.calls {
				t.Fatalf("repository calls = %d, want %d", repo.findGameCalls, tt.calls)
			}
			if tt.calls > 0 && repo.gameUUID != gameUUID {
				t.Fatal("wrong game UUID passed to repository")
			}
			if tt.status != http.StatusOK {
				if strings.Contains(response.Body.String(), "private database details") {
					t.Fatal("database details exposed")
				}
				if strings.Contains(response.Body.String(), `"table"`) {
					t.Fatal("game returned with error")
				}
				return
			}
			var got GameDetailsResponse
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.UUID != gameUUID || got.State != tt.state.String() || got.Table != game.Table || got.FirstPlayerUUID != first || got.SecondPlayerUUID != second {
				t.Fatalf("unexpected game response: %+v", got)
			}
		})
	}
}

func TestGameHandlerGetGames(t *testing.T) {
	const playerUUID = "11111111-1111-4111-8111-111111111111"
	game := domain.GameInstance{UUID: "22222222-2222-4222-8222-222222222222", FirstPlayerUUID: "33333333-3333-4333-8333-333333333333", State: domain.WaitingForPlayers}
	for _, tt := range []struct {
		name   string
		player string
		games  []domain.GameInstance
		err    error
		status int
	}{
		{name: "available games", player: playerUUID, games: []domain.GameInstance{game}, status: http.StatusOK},
		{name: "empty list", player: playerUUID, status: http.StatusOK},
		{name: "unauthenticated", status: http.StatusUnauthorized},
		{name: "database failure", player: playerUUID, err: errors.New("private database details"), status: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &gameRepositoryStub{availableGames: tt.games, listGamesErr: tt.err}
			handler := NewGameHandler(services.NewGameService(repo, zap.NewNop()), zap.NewNop())
			request := httptest.NewRequest(http.MethodGet, "/game", nil)
			if tt.player != "" {
				request = request.WithContext(requestcontext.WithUserUUID(request.Context(), tt.player))
			}
			response := httptest.NewRecorder()
			handler.GetGames(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body.String())
			}
			if tt.player == "" {
				if repo.listGamesCalls != 0 {
					t.Fatal("repository called without authentication")
				}
				return
			}
			if repo.playerUUID != tt.player {
				t.Fatal("wrong player UUID passed to repository")
			}
			if tt.err != nil {
				if response.Body.String() != "internal server error\n" {
					t.Fatal("database details exposed")
				}
				return
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("missing JSON content type")
			}
			var got []GameDetailsResponse
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.games) {
				t.Fatalf("got %d games, want %d", len(got), len(tt.games))
			}
			if len(got) > 0 && (got[0].UUID != game.UUID || got[0].State != "waiting_for_players" || got[0].FirstPlayerUUID != game.FirstPlayerUUID) {
				t.Fatalf("unexpected game: %+v", got[0])
			}
			if len(tt.games) == 0 && strings.TrimSpace(response.Body.String()) != "[]" {
				t.Fatal("empty list must be [], not null")
			}
		})
	}
}
