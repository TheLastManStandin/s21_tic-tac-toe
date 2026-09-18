package middleware

import (
	"net/http"
	"strings"
)

func BasicAuthWrap(handler http.Handler) http.Handler {
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
		handler.ServeHTTP(w, r)
	})
}
