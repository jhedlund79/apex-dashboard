package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"apex-dashboard/backend/internal/config"
	"apex-dashboard/backend/internal/handlers"
	"apex-dashboard/backend/internal/repository/cache"
	"apex-dashboard/backend/internal/repository/live"
	plaidrepo "apex-dashboard/backend/internal/repository/plaid"
	simrepo "apex-dashboard/backend/internal/repository/simulation"
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
	if cfg.CoinGeckoKey == "" {
		logger.Error("COINGECKO_API_KEY not set — CoinGecko API will be disabled")
		os.Exit(1)
	}

	liveStore, err := live.New(cfg.PolygonKey, cfg.CoinGeckoKey)
	if err != nil {
		logger.Error("failed to create live store", "err", err)
		os.Exit(1)
	}

	// Cache results for 60 seconds to respect free-tier API rate limits.
	cachedStore := cache.New(liveStore, 60*time.Second)
	logger.Info("live market data enabled", "cache_ttl", "60s")

	simStore, err := simrepo.New(cfg.SimulationDataFile)
	if err != nil {
		logger.Error("failed to create simulation store", "err", err)
		os.Exit(1)
	}
	logger.Info("simulation store ready", "file", cfg.SimulationDataFile)

	handlerCfg := handlers.Config{
		Store:      cachedStore,
		Logger:     logger,
		Simulation: simStore,
		Quoter:     liveStore,
	}

	if cfg.PlaidClientID != "" && cfg.PlaidSecret != "" {
		ps, err := plaidrepo.New(plaidrepo.Config{
			Inner:     cachedStore,
			ClientID:  cfg.PlaidClientID,
			Secret:    cfg.PlaidSecret,
			Env:       cfg.PlaidEnv,
			TokenFile: "./plaid-tokens.json",
		})
		if err != nil {
			logger.Error("failed to create Plaid store", "err", err)
			os.Exit(1)
		}
		handlerCfg.Store = ps
		handlerCfg.Plaid = ps
		logger.Info("Plaid integration enabled", "env", cfg.PlaidEnv)
	} else {
		logger.Info("Plaid not configured — portfolio pages will use seed data")
	}

	h, err := handlers.New(handlerCfg)
	if err != nil {
		logger.Error("failed to create handlers", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/overview", h.HandleOverview)
	mux.HandleFunc("GET /api/v1/markets", h.HandleMarkets)
	mux.HandleFunc("GET /api/v1/crypto", h.HandleCrypto)
	mux.HandleFunc("GET /api/v1/recommendations", h.HandleRecommendations)
	mux.HandleFunc("GET /api/v1/portfolio/principal", h.HandlePrincipal)
	mux.HandleFunc("GET /api/v1/portfolio/morgan-stanley", h.HandleMorganStanley)
	mux.HandleFunc("GET /api/v1/portfolio/fidelity", h.HandleFidelity)
	mux.HandleFunc("GET /api/v1/portfolio/sofi", h.HandleSoFi)
	mux.HandleFunc("GET /api/v1/plaid/status", h.HandlePlaidStatus)
	mux.HandleFunc("POST /api/v1/plaid/link-token", h.HandlePlaidLinkToken)
	mux.HandleFunc("POST /api/v1/plaid/exchange", h.HandlePlaidExchange)
	mux.HandleFunc("GET /api/v1/simulation", h.HandleSimulation)
	mux.HandleFunc("GET /api/v1/simulation/quote", h.HandleSimulationQuote)
	mux.HandleFunc("POST /api/v1/simulation/trade", h.HandleSimulationTrade)
	mux.HandleFunc("POST /api/v1/simulation/reset", h.HandleSimulationReset)

	// Serve static files from frontend/dist for SPA
	distDir := "./frontend/dist"
	fs := http.FileServer(http.Dir(distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Check if file exists in dist
		path := distDir + r.URL.Path
		if _, err := os.Stat(path); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		// Serve index.html for SPA routes
		http.ServeFile(w, r, distDir+"/index.html")
	})

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
