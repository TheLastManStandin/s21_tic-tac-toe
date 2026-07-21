package di

import (
	"datasource"
	"domain"
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
				//datasource.NewMapDatabase,
				datasource.NewPostgresDatabase,
				fx.As(new(datasource.Database))),
			domain.NewGameService,
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
