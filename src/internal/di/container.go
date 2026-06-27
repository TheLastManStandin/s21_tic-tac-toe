package di

import (
	"api"
	"fmt"
	"net/http"

	"go.uber.org/fx"
)

func Build() {
	fx.New(
		fx.Provide(
			api.InitHTTPServer,
		),
		fx.Invoke(func(srv *http.Server) {
			fmt.Println("starting server")
		}),
	).Run()
}
