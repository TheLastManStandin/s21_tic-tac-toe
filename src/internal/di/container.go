package di

import (
	"api"
	"datasource"
	"domain"
	"fmt"
	"net/http"

	"go.uber.org/fx"
)

func Build() {
	fx.New(
		fx.Provide(
			datasource.NewGamesDatabase,
			datasource.NewGameRepository,
			fx.Annotate(
				domain.NewService,
				fx.As(new(domain.GameService)),
			),
			api.NewHandler,
			api.InitHTTPServer,
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
