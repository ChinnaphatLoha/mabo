package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
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
	netCfg := network.Config{BindAddress: cfg.BindAddress}
	networkServer := network.NewUDPServer(netCfg, nil, log)

	var sender app.Sender = networkServer

	latencyMs := 0
	if val := os.Getenv("LATENCY_MS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			latencyMs = parsed
			log.Info("network latency simulation enabled", "latency_ms", latencyMs)
		}
	}

	if latencyMs > 0 {
		sender = network.NewSimulatedSender(networkServer, latencyMs)
	}

	// ── Application handler ────────────────────────────────────────────────
	handler := app.NewMultiplayerHandler(
		sessions, rooms, matches, w, inputs,
		sender, log,
	)

	var pktHandler network.PacketHandler = handler
	if latencyMs > 0 {
		pktHandler = network.NewSimulatedHandler(handler, latencyMs)
	}

	// Set the handler on the already-created network server.
	networkServer.SetHandler(pktHandler)

	// ── Interest manager & tick loop ───────────────────────────────────────
	interest := system.BroadcastInterestManager{}
	tickLoop := app.NewTickLoop(sessions, matches, w, inputs, interest, sender, log)

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
