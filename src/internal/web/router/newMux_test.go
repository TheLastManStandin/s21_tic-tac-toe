package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/repositorys"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"

	"go.uber.org/zap"
)

type readRepositoryStub struct {
	repositorys.GameRepository
	repositorys.UserRepository
}

const testUserUUID = "11111111-1111-4111-8111-111111111111"
const testGameUUID = "22222222-2222-4222-8222-222222222222"

func (readRepositoryStub) ListAvailableGames(context.Context, string) ([]domain.GameInstance, error) {
	return []domain.GameInstance{}, nil
}

func (readRepositoryStub) FindGameByID(context.Context, string) (domain.GameInstance, error) {
	return domain.GameInstance{UUID: testGameUUID, FirstPlayerUUID: testUserUUID}, nil
}

func (readRepositoryStub) FindUserByID(context.Context, string) (domain.UserInstance, error) {
	return domain.UserInstance{UUID: testUserUUID, Username: "player"}, nil
}

type authenticatorStub struct{}

func (authenticatorStub) AuthenticateUser(_ context.Context, username, password string) (string, bool) {
	return testUserUUID, username == "player" && password == "secret"
}

func TestReadRoutesRequireAuthentication(t *testing.T) {
	repo := readRepositoryStub{}
	logger := zap.NewNop()
	users := services.NewUserService(repo, logger)
	mux := NewServeMux(
		handlers.NewGameHandler(services.NewGameService(repo, logger), logger),
		handlers.NewAuthHandler(users, logger),
		handlers.NewUserHandler(users, logger),
		middleware.NewBasicAuth(authenticatorStub{}),
	)
	for _, path := range []string{"/game", "/game/" + testGameUUID, "/user", "/user?UUID=" + testUserUUID} {
		for _, authenticated := range []bool{false, true} {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			wantStatus := http.StatusUnauthorized
			if authenticated {
				request.SetBasicAuth("player", "secret")
				wantStatus = http.StatusOK
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != wantStatus {
				t.Fatalf("GET %s (authenticated=%t): status %d, want %d: %s", path, authenticated, response.Code, wantStatus, response.Body.String())
			}
			if authenticated && response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("GET %s did not return JSON", path)
			}
		}
	}
}
