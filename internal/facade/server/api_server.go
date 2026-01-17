package server

import (
	"context"
	"errors"
	"kiwi-user/config"
	"kiwi-user/internal/facade/server/route"
	"net/http"

	libgin "github.com/Yet-Another-AI-Project/kiwi-lib/server/gin"
	"github.com/futurxlab/golanggraph/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type APIServer struct {
	Engine *gin.Engine
}

func NewAPIServer(lc fx.Lifecycle,
	logger logger.ILogger,
	cfg *config.Config,
	route *route.Route) (*APIServer, error) {

	engine, err := libgin.NewGin(
		libgin.WithLogger(logger),
		libgin.WithServiceName("kiwi-user"),
		libgin.WithMetricsEndpoint(cfg.Metrics.MetricsEndpoint),
	)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    ":" + cfg.APIServer.Port,
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Infof(ctx, "api server start listening, port %s", cfg.APIServer.Port)
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Errorf(ctx, "API server listen error %w", err)
				}
			}()
			return nil
		},

		OnStop: func(ctx context.Context) error {
			logger.Infof(ctx, "api server shutdown...")
			if err := srv.Shutdown(ctx); err != nil {
				logger.Errorf(ctx, "api server shutdown error %w", err)
			}
			return nil
		},
	})

	return &APIServer{
		Engine: engine,
	}, nil
}
