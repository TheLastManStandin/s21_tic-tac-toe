package middleware

import (
	"context"
	"net/http"
	"strings"
	"tickTackToe/internal/web/requestcontext"
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

// todo переписать как нормальная мидлваря
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

		uuid, ok := a.authenticator.AuthenticateUser(
			r.Context(),
			login,
			password,
		)

		if ok == false {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := requestcontext.WithUserUUID(r.Context(), uuid)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
