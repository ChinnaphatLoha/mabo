package app

import (
	"context"

	"github.com/ChinnaphatLoha/mabo/server/internal/config"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/network"
)

type App struct {
	config  *config.Config
	logger  *logger.Logger
	network network.Server
}

func New(cfg *config.Config, logger *logger.Logger, networkServer network.Server) *App {
	return &App{
		config:  cfg,
		logger:  logger,
		network: networkServer,
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.network.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutting down application")
		return a.network.Stop(ctx)
	case err := <-errCh:
		return err
	}
}
