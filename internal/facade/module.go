package facade

import (
	"kiwi-user/internal/facade/controller/admin"
	"kiwi-user/internal/facade/controller/api"
	"kiwi-user/internal/facade/server"
	"kiwi-user/internal/facade/server/route"

	"go.uber.org/fx"
)

var APIServerModule = fx.Provide(
	api.NewController,
	admin.NewController,
	route.NewRoute,
	server.NewAPIServer,
)
