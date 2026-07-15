package di

import (
	"datasource"
	"fmt"
	"net/http"
	"web"

	"go.uber.org/fx"
)

func Build() {
	fx.New(
		fx.Provide(
			web.InitHTTPServer,
			fx.Annotate(
				datasource.NewGamesDatabase,
				fx.As(new(datasource.Database))),
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
