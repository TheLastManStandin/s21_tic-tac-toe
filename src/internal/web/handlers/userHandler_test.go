package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/requestcontext"

	"go.uber.org/zap"
)

func TestUserHandlerGetUserStatus(t *testing.T) {
	const current = "11111111-1111-4111-8111-111111111111"
	const other = "22222222-2222-4222-8222-222222222222"
	for _, tt := range []struct {
		name     string
		url      string
		player   string
		wantUUID string
		err      error
		status   int
	}{
		{name: "current user", url: "/user", player: current, wantUUID: current, status: http.StatusOK},
		{name: "user by UUID", url: "/user?UUID=" + other, player: current, wantUUID: other, status: http.StatusOK},
		{name: "missing user", url: "/user?UUID=" + other, player: current, wantUUID: other, err: fmt.Errorf("repository: %w", domain.ErrUserNotFound), status: http.StatusNotFound},
		{name: "database failure", url: "/user", player: current, wantUUID: current, err: errors.New("private database details"), status: http.StatusInternalServerError},
		{name: "unauthenticated", url: "/user", status: http.StatusUnauthorized},
		{name: "invalid UUID", url: "/user?UUID=invalid", player: current, status: http.StatusBadRequest},
		{name: "empty UUID", url: "/user?UUID=", player: current, status: http.StatusBadRequest},
		{name: "duplicate UUID", url: "/user?UUID=" + current + "&UUID=" + other, player: current, status: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &userRepositoryStub{foundUser: domain.UserInstance{UUID: tt.wantUUID, Username: "player", Password: "private password hash"}, findErr: tt.err}
			handler := NewUserHandler(services.NewUserService(repo, zap.NewNop()), zap.NewNop())
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			if tt.player != "" {
				request = request.WithContext(requestcontext.WithUserUUID(request.Context(), tt.player))
			}
			response := httptest.NewRecorder()
			handler.GetUserStatus(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body.String())
			}
			if tt.wantUUID == "" && repo.findCalls != 0 {
				t.Fatal("invalid request reached repository")
			}
			if tt.wantUUID != "" && (repo.findCalls != 1 || repo.requestedUUID != tt.wantUUID) {
				t.Fatalf("wrong lookup: %d calls, UUID %q", repo.findCalls, repo.requestedUUID)
			}
			if tt.status == http.StatusInternalServerError && response.Body.String() != "internal server error\n" {
				t.Fatal("database details exposed")
			}
			if tt.status != http.StatusOK {
				return
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 2 || body["uuid"] != tt.wantUUID || body["username"] != "player" {
				t.Fatalf("response must contain only UUID and username: %v", body)
			}
		})
	}
}
