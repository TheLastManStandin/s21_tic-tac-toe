package router

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

func NewServeMux(
	gameService domain.GameService,
	log *zap.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /game/{UUID}", GamePostHandler{
		gameService: gameService,
		log:         log,
	})
	//mux.HandleFunc("POST /registration", registrationHandler(db))
	//mux.HandleFunc("POST /authorization", authenticationMiddleware(db, uuidRegexp, authHandler))
	return mux
}

func validateRequest(req GameRequest) error {
	if req.Y < 0 || req.X < 0 || req.Y > 2 || req.X > 2 {
		return errors.New("invalid cordinates")
	}

	return nil
}
