package di

import (
	"fmt"
	"net/http"
	"tickTackToe/internal/datasource"
	"tickTackToe/internal/domain/gameRepository"
	"tickTackToe/internal/domain/gameService"
	"tickTackToe/internal/migrations"
	"tickTackToe/internal/web"
	"tickTackToe/internal/web/handlers"
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
			gameService.NewGameService,
			fx.Annotate(
				datasource.NewPostgresDatabase,
				fx.As(new(gameRepository.GameRepository)),
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
