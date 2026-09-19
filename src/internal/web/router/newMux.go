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

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("GET /auth", authHandler.Auth)

	mux.Handle("/", auth.Wrap(protectedMux))

	return mux
}
