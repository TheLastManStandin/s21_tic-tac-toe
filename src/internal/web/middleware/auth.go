package middleware

// func AuthMiddleware(ua web.UserAuthenticator) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			uuid, ok := ua.AuthenticateUser(r)

// 			if !ok || uuid == "" {
// 				httpresponse.WriteUnauthorized(w)
// 				return
// 			}

// 			ctx := requestcontext.WithUserUUID(r.Context(), uuid)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }
