package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"tickTackToe/internal/web/requestcontext"
)

type userAuthenticatorFunc func(
	ctx context.Context,
	username string,
	password string,
) (string, bool)

func (f userAuthenticatorFunc) AuthenticateUser(
	ctx context.Context,
	username string,
	password string,
) (string, bool) {
	return f(ctx, username, password)
}

func TestBasicAuthWrap(t *testing.T) {
	t.Run("rejects request without credentials", func(t *testing.T) {
		nextCalled := false
		authenticatorCalled := false
		auth := NewBasicAuth(userAuthenticatorFunc(func(
			context.Context,
			string,
			string,
		) (string, bool) {
			authenticatorCalled = true
			return "", false
		}))
		handler := auth.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			nextCalled = true
		}))

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/game/id", nil))

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
		if response.Header().Get("WWW-Authenticate") == "" {
			t.Fatal("WWW-Authenticate header is missing")
		}
		if authenticatorCalled {
			t.Fatal("authenticator was called without credentials")
		}
		if nextCalled {
			t.Fatal("protected handler was called")
		}
	})

	t.Run("rejects empty credentials", func(t *testing.T) {
		authenticatorCalled := false
		auth := NewBasicAuth(userAuthenticatorFunc(func(
			context.Context,
			string,
			string,
		) (string, bool) {
			authenticatorCalled = true
			return "", false
		}))
		handler := auth.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("protected handler was called")
		}))
		request := httptest.NewRequest(http.MethodGet, "/game/id", nil)
		request.SetBasicAuth("", "")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
		if authenticatorCalled {
			t.Fatal("authenticator was called with empty credentials")
		}
	})

	t.Run("rejects invalid credentials", func(t *testing.T) {
		auth := NewBasicAuth(userAuthenticatorFunc(func(
			_ context.Context,
			username string,
			password string,
		) (string, bool) {
			if username != "player" || password != "wrong-secret" {
				t.Fatalf("credentials = %q:%q, want player:wrong-secret", username, password)
			}
			return "", false
		}))
		handler := auth.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("protected handler was called")
		}))
		request := httptest.NewRequest(http.MethodGet, "/game/id", nil)
		request.SetBasicAuth("player", "wrong-secret")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
		if response.Header().Get("WWW-Authenticate") == "" {
			t.Fatal("WWW-Authenticate header is missing")
		}
		if response.Body.String() != "unauthorized\n" {
			t.Fatalf("body = %q, want unauthorized", response.Body.String())
		}
	})

	t.Run("passes authenticated request with user UUID", func(t *testing.T) {
		const wantUUID = "11111111-1111-4111-8111-111111111111"

		auth := NewBasicAuth(userAuthenticatorFunc(func(
			_ context.Context,
			username string,
			password string,
		) (string, bool) {
			if username != "player" || password != "secret" {
				t.Fatalf("credentials = %q:%q, want player:secret", username, password)
			}
			return wantUUID, true
		}))
		handler := auth.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userUUID, ok := requestcontext.UserUUID(r.Context())
			if !ok {
				t.Fatal("user UUID is missing from request context")
			}
			if userUUID != wantUUID {
				t.Fatalf("user UUID = %q, want %q", userUUID, wantUUID)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		request := httptest.NewRequest(http.MethodGet, "/game/id", nil)
		request.SetBasicAuth("player", "secret")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
	})
}
