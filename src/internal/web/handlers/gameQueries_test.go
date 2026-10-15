package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/requestcontext"

	"go.uber.org/zap"
)

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
