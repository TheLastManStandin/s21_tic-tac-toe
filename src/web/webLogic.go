package web

import (
	"context"
	"datasource"
	"domain"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"go.uber.org/fx"
)

func InitHTTPServer(
	lc fx.Lifecycle,
	gameService domain.GameService,
	db datasource.Database,
) *http.Server {
	mux := http.NewServeMux()
	uuidRegexp := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	//TODO добавить контекст
	mux.HandleFunc("POST /game/{UUID}", authenticationMiddleware(db, uuidRegexp, gamePostHandler))
	mux.HandleFunc("POST /registration", registrationHandler(db))
	mux.HandleFunc("POST /authorization", authenticationMiddleware(db, uuidRegexp, authHandler))

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
			fmt.Println("SHUTDOWN")
			err := db.Close()
			if err != nil {
				fmt.Println("DB CLOSE ERROR")
			}
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func validateRequest(req GameRequest, UUID string, uuidRegexp regexp.Regexp) error {
	if !uuidRegexp.MatchString(UUID) {
		return errors.New("invalid uuid")
	}

	if req.Y < 0 || req.X < 0 || req.Y > 2 || req.X > 2 {
		return errors.New("invalid cordinates")
	}

	return nil
}
