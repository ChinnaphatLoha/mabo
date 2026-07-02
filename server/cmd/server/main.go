package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChinnaphatLoha/mabo/server/internal/app"
	"github.com/ChinnaphatLoha/mabo/server/internal/config"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/network"
)

func main() {
	cfg, err := config.Load("configs/server.yml")
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel)
	log.Info("starting server", "bind_address", cfg.BindAddress, "port", cfg.Port)

	networkServer := network.NewUDPServer(network.Config{
		BindAddress: cfg.BindAddress,
	}, nil, log)

	app := app.New(cfg, log, networkServer)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.Error("server error", "error", err.Error())
		os.Exit(1)
	}

	log.Info("server stopped")
}
