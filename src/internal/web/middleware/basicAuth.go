package middleware

import (
	"context"
	"net/http"
	"strings"
)

type UserAuthenticator interface {
	AuthenticateUser(
		ctx context.Context,
		username string,
		password string,
	) (string, bool)
}

type BasicAuth struct {
	authenticator UserAuthenticator
}

func NewBasicAuth(authenticator UserAuthenticator) *BasicAuth {
	return &BasicAuth{authenticator: authenticator}
}

func (a *BasicAuth) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, password, ok := r.BasicAuth()
		if !ok || strings.TrimSpace(login) == "" || password == "" {
			w.Header().Set(
				"WWW-Authenticate",
				`Basic realm="tic-tac-toe", charset="UTF-8"`,
			)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Validate login and password through AuthService and put the
		// authenticated user's UUID into the request context.
		_, ok = a.authenticator.AuthenticateUser(r.Context(), login, password)
		if ok == false {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
