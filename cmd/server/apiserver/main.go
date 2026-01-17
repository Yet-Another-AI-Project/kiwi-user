package main

import (
	"context"
	"kiwi-user/config"
	"kiwi-user/internal/application"
	"kiwi-user/internal/bootstrap"
	"kiwi-user/internal/domain"
	"kiwi-user/internal/facade"
	"kiwi-user/internal/facade/server"
	"kiwi-user/internal/facade/server/route"
	"kiwi-user/internal/infrastructure"
	"kiwi-user/internal/infrastructure/jwt"
	"log"
	"os"
	"os/signal"
	"time"

	"go.uber.org/fx"

	_ "github.com/lib/pq"
)

func registerRoute(route *route.Route, apiserver *server.APIServer) error {
	route.RegisterApiV1(apiserver.Engine)
	route.RegisterAdmin(apiserver.Engine)
	return nil
}

func initJWT(rsa *jwt.RSA) error {
	return rsa.Init()
}

func initBootstrap(b *bootstrap.Bootstrap) error {
	return b.Init()
}

func main() {

	app := fx.New(
		fx.Provide(
			config.NewConfig,
		),
		facade.APIServerModule,
		application.Module,
		domain.Module,
		bootstrap.Module,
		infrastructure.Module,
		fx.NopLogger,
		fx.Invoke(registerRoute),
		fx.Invoke(initJWT),
		fx.Invoke(initBootstrap),
	)
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}
