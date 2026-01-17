package route

import (
	"kiwi-user/config"
	"kiwi-user/internal/facade/controller/admin"
	"kiwi-user/internal/facade/controller/api"
	"kiwi-user/internal/infrastructure/jwt"

	"github.com/futurxlab/golanggraph/logger"
)

type Route struct {
	config          *config.Config
	apiController   *api.Controller
	adminController *admin.Controller
	logger          logger.ILogger
	jwtHepler       *jwt.JWTHelper
}

func NewRoute(
	config *config.Config,
	apiController *api.Controller,
	adminController *admin.Controller,
	logger logger.ILogger,
	jwtHepler *jwt.JWTHelper) *Route {

	return &Route{
		config:          config,
		apiController:   apiController,
		adminController: adminController,
		logger:          logger,
		jwtHepler:       jwtHepler,
	}
}
