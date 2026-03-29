package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apex-dashboard/backend/internal/config"
	"apex-dashboard/backend/internal/handlers"
	"apex-dashboard/backend/internal/repository/cache"
	"apex-dashboard/backend/internal/repository/live"
	"apex-dashboard/backend/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()

	if cfg.PolygonKey == "" {
		logger.Error("POLYGON_API_KEY is required — set it in .env or the environment")
		os.Exit(1)
	}

	liveStore, err := live.New(cfg.PolygonKey, cfg.CoinGeckoKey)
	if err != nil {
		logger.Error("failed to create live store", "err", err)
		os.Exit(1)
	}

	// Cache results for 60 seconds to respect free-tier API rate limits.
	store := cache.New(liveStore, 60*time.Second)
	logger.Info("live market data enabled", "cache_ttl", "60s")

	h, err := handlers.New(handlers.Config{
		Store:  store,
		Logger: logger,
	})
	if err != nil {
		logger.Error("failed to create handlers", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/overview", h.HandleOverview)
	mux.HandleFunc("GET /api/v1/markets", h.HandleMarkets)
	mux.HandleFunc("GET /api/v1/crypto", h.HandleCrypto)
	mux.HandleFunc("GET /api/v1/recommendations", h.HandleRecommendations)

	srv, err := server.New(server.Config{
		Addr:    cfg.Addr,
		Handler: mux,
		Logger:  logger,
	})
	if err != nil {
		logger.Error("failed to create server", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
