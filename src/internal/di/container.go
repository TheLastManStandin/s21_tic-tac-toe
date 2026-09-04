package di

import (
	"fmt"
	"internal/internal/datasource"
	"internal/internal/domain"
	"internal/internal/web"
	"net/http"
	"tickTackToe/internal/web"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Build() {
	fx.New(
		fx.Provide(
			web.InitHTTPServer,
			fx.Annotate(
				datasource.NewPostgresDatabase,
				fx.As(new(datasource.Database))),
			domain.NewGameService,
			domain.NewUserService,
			web.NewServeMux,
			zap.NewDevelopment,
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
