package middleware

import (
	"net/http"
	"strings"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/web/utils/httpresponse"
	"tickTackToe/internal/web/utils/requestcontext"
)

func AuthMiddleware(ua domain.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpresponse.WriteUnauthorized(w)
				return
			}

			userUUID, err := ua.AuthenticateAccessToken(r.Context(), parts[1])
			if err != nil {
				httpresponse.WriteUnauthorized(w)
				return
			}

			if err != nil || userUUID == "" {
				httpresponse.WriteUnauthorized(w)
				return
			}

			ctx := requestcontext.WithUserUUID(r.Context(), userUUID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
