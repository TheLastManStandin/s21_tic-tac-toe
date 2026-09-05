package di

import (
	"fmt"
	"net/http"
	"tickTackToe/internal/web"
	"tickTackToe/internal/web/router"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Build() {
	fx.New(
		fx.Provide(
			web.InitHTTPServer,
			router.NewServeMux,
			zap.NewDevelopment,
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
