package web

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func InitHTTPServer(
	lc fx.Lifecycle,
	log *zap.Logger,
	mux *http.ServeMux,
) *http.Server {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":5005"
	}

	srv := &http.Server{
		Addr:    serverAddress,
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
