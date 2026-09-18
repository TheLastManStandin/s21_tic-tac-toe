package router

import (
	"net/http"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"
)

func NewServeMux(
	gameHandler *handlers.GameHandler,
) *http.ServeMux {
	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("POST /game/{UUID}", gameHandler.MakeTurn)

	mux.Handle("/", middleware.BasicAuthWrap(protectedMux))

	return mux
}
