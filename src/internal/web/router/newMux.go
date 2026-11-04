package router

import (
	"net/http"
	"tickTackToe/internal/web/handlers"
)

func NewServeMux(
	gameHandler *handlers.GameHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	// auth *middleware.BasicAuth,
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

	// mux.Handle("/", auth.Wrap(protectedMux))
	mux.Handle("/", addMiddleware(
		protectedMux,
		// middleware.BasicAuthMiddleware(auth),
	))

	return mux
}

func addMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, mv := range middlewares {
		handler = mv(handler)
	}
	return handler
}
