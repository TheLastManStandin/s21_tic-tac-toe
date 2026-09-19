package di

import (
	"fmt"
	"net/http"
	"tickTackToe/internal/datasource"
	"tickTackToe/internal/domain/repositorys"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/migrations"
	"tickTackToe/internal/web"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"
	"tickTackToe/internal/web/router"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Build() {
	fx.New(
		fx.Provide(
			web.InitHTTPServer,

			zap.NewDevelopment,

			router.NewServeMux,

			handlers.NewGameHandler,
			services.NewGameService,

			handlers.NewAuthHandler,
			services.NewUserService,

			middleware.NewBasicAuth,

			fx.Annotate(
				datasource.NewPostgresDatabase,
				fx.As(new(repositorys.GameRepository)),
				fx.As(new(repositorys.UserRepository)),
			),
			fx.Annotate(
				services.NewUserService,
				fx.As(new(middleware.UserAuthenticator)),
			),
		),
		fx.Invoke(
			migrations.Up,
			func(srv *http.Server) {
				fmt.Println("starting server")
			},
		),
	).Run()
}
