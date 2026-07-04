package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChinnaphatLoha/mabo/server/internal/app"
	"github.com/ChinnaphatLoha/mabo/server/internal/command"
	"github.com/ChinnaphatLoha/mabo/server/internal/config"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/match"
	"github.com/ChinnaphatLoha/mabo/server/internal/network"
	"github.com/ChinnaphatLoha/mabo/server/internal/room"
	"github.com/ChinnaphatLoha/mabo/server/internal/session"
	"github.com/ChinnaphatLoha/mabo/server/internal/system"
	"github.com/ChinnaphatLoha/mabo/server/internal/world"
)

func main() {
	cfg, err := config.Load("configs/server.yml")
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel)
	log.Info("starting server", "bind_address", cfg.BindAddress, "port", cfg.Port)

	// ── Domain managers ────────────────────────────────────────────────────
	sessions := session.NewManager()
	rooms := room.NewManager()
	matches := match.NewManager()
	w := world.New()
	inputs := command.NewInputBuffer()

	// ── Network ────────────────────────────────────────────────────────────
	// The UDP server is created first so it can be passed as the Sender.
	// The handler is wired after so it has a reference to the live server.
	netCfg := network.Config{BindAddress: cfg.BindAddress}
	// Placeholder handler lets us pass the server to the multiplayer handler.
	// The real handler is set below once all managers are constructed.
	networkServer := network.NewUDPServer(netCfg, nil, log)

	// ── Application handler ────────────────────────────────────────────────
	handler := app.NewMultiplayerHandler(
		sessions, rooms, matches, w, inputs,
		networkServer, log,
	)

	// Re-create network server with the real handler.
	networkServer = network.NewUDPServer(netCfg, handler, log)

	// ── Interest manager & tick loop ───────────────────────────────────────
	interest := system.BroadcastInterestManager{}
	tickLoop := app.NewTickLoop(sessions, matches, w, inputs, interest, networkServer, log)

	// ── Application lifecycle ──────────────────────────────────────────────
	application := app.New(cfg, log, networkServer)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Run tick loop in background.
	go tickLoop.Run(ctx)

	if err := application.Run(ctx); err != nil {
		log.Error("server error", "error", err.Error())
		os.Exit(1)
	}

	log.Info("server stopped")
}
