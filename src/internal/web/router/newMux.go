package router

import (
	"net/http"
	"tickTackToe/internal/web/handlers"
)

func NewServeMux(
	gameHandler *handlers.GameHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /game/{UUID}", gameHandler.MakeTurn)

	return mux
}
