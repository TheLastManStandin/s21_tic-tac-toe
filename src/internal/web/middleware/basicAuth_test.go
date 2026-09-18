package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthWrap(t *testing.T) {
	t.Run("rejects request without credentials", func(t *testing.T) {
		nextCalled := false
		handler := BasicAuthWrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
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
		if nextCalled {
			t.Fatal("protected handler was called")
		}
	})

	t.Run("rejects empty credentials", func(t *testing.T) {
		handler := BasicAuthWrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("protected handler was called")
		}))
		request := httptest.NewRequest(http.MethodGet, "/game/id", nil)
		request.SetBasicAuth("", "")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("passes request with credentials", func(t *testing.T) {
		nextCalled := false
		handler := BasicAuthWrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		}))
		request := httptest.NewRequest(http.MethodGet, "/game/id", nil)
		request.SetBasicAuth("player", "secret")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
		if !nextCalled {
			t.Fatal("protected handler was not called")
		}
	})
}
