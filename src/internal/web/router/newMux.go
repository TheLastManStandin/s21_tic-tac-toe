package router

import (
	"net/http"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"
)

func NewServeMux(
	gameHandler *handlers.GameHandler,
	authHandler *handlers.AuthHandler,
	auth *middleware.BasicAuth,
) *http.ServeMux {
	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("POST /game/{UUID}", gameHandler.MakeTurn)
	// todo ↓
	protectedMux.HandleFunc("POST /game", gameHandler.CreateGame)
	//protectedMux.HandleFunc("GET /game", gameHandler.GetGames)
	//protectedMux.HandleFunc("POST /game/{UUID}/join", gameHandler.JoinGame)
	//protectedMux.HandleFunc("GET /game/{UUID}", gameHandler.GetGameStatus)
	//protectedMux.HandleFunc("GET /user", userHandler.GetUserStatus)

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("GET /auth", authHandler.Auth)

	mux.Handle("/", auth.Wrap(protectedMux))

	return mux
}
