package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"

	"go.uber.org/zap"
)

type userRepositoryStub struct {
	insertErr   error
	insertCalls int
}

func (r *userRepositoryStub) GetUser(context.Context, string) (domain.UserInstance, error) {
	return domain.UserInstance{}, nil
}

func (r *userRepositoryStub) InsertUser(context.Context, domain.UserInstance) error {
	r.insertCalls++
	return r.insertErr
}

func TestAuthHandlerRegisterErrors(t *testing.T) {
	for _, tt := range []struct {
		name        string
		credentials bool
		err         error
		wantStatus  int
		wantBody    string
	}{
		{name: "missing credentials", wantStatus: http.StatusUnauthorized, wantBody: "unauthorized\n"},
		{name: "existing user", credentials: true, err: fmt.Errorf("insert user: %w", domain.ErrUserAlreadyExists), wantStatus: http.StatusConflict, wantBody: "user already exists\n"},
		{name: "database error", credentials: true, err: errors.New("private database details"), wantStatus: http.StatusInternalServerError, wantBody: "internal server error\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &userRepositoryStub{insertErr: tt.err}
			handler := NewAuthHandler(services.NewUserService(repo, zap.NewNop()), zap.NewNop())
			request := httptest.NewRequest(http.MethodPost, "/register", nil)
			if tt.credentials {
				request.SetBasicAuth("player", "secret")
			}
			response := httptest.NewRecorder()
			handler.Register(response, request)
			if response.Code != tt.wantStatus || response.Body.String() != tt.wantBody {
				t.Fatalf("response = (%d, %q), want (%d, %q)", response.Code, response.Body.String(), tt.wantStatus, tt.wantBody)
			}
			if !tt.credentials {
				if repo.insertCalls != 0 {
					t.Fatal("user was inserted without credentials")
				}
				if response.Header().Get("WWW-Authenticate") == "" {
					t.Fatal("WWW-Authenticate header is missing")
				}
			}
		})
	}
}
