package web

import (
	"context"
	"datasource"
	"domain"
	"errors"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func InitHTTPServer(
	lc fx.Lifecycle,
	db datasource.Database,
	log *zap.Logger,
	mux *http.ServeMux,
) *http.Server {
	srv := &http.Server{
		Addr:    ":5005",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					return
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("SHUTDOWN")
			err := db.Close()
			if err != nil {
				log.Info("DB CLOSE ERROR")
			}
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

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
