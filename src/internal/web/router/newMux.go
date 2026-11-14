package router

import (
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"
)

func NewServeMux(
	gameHandler *handlers.GameHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	auth domain.AuthService,
) *http.ServeMux {
	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("POST /game/{UUID}", gameHandler.MakeTurn)
	protectedMux.HandleFunc("POST /game", gameHandler.CreateGame)
	protectedMux.HandleFunc("POST /game/{UUID}/join", gameHandler.JoinGame)
	protectedMux.HandleFunc("GET /game", gameHandler.GetGames)
	protectedMux.HandleFunc("GET /game/{UUID}", gameHandler.GetGameStatus)
	protectedMux.HandleFunc("GET /user/{UUID}", userHandler.GetUserStatus)

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("GET /auth", authHandler.Auth)
	mux.HandleFunc("/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// mux.Handle("/", auth.Wrap(protectedMux))
	mux.Handle("/", addMiddleware(
		protectedMux,
		middleware.AuthMiddleware(auth),
	))

	return mux
}

func addMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, mv := range middlewares {
		handler = mv(handler)
	}
	return handler
}
