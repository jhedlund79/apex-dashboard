// Package server provides the HTTP API server for the APEX dashboard.
//
// The primary entrypoint is [New], which validates [Config], wires routes
// with middleware, and returns a *Server. Call [Server.Run] with a context
// to start serving; Run blocks until the context is cancelled and then
// drains in-flight requests before returning.
//
// Routes served:
//
//	GET /api/v1/overview        — overview page data
//	GET /api/v1/markets         — US + international market data
//	GET /api/v1/crypto          — crypto prices and history
//	GET /api/v1/recommendations — growth pick recommendations
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"apex-dashboard/backend/internal/middleware"
)

// shutdownTimeout is the maximum time allowed to drain in-flight requests
// after the context is cancelled.
const shutdownTimeout = 10 * time.Second

// Config controls Server construction and runtime behavior.
// Addr defaults to ":8080" when empty. Handler must not be nil.
type Config struct {
	// Required: the root http.Handler (typically a mux with routes registered).
	Handler http.Handler

	// Optional: listen address. Defaults to ":8080".
	Addr string

	// Optional: nil disables structured logging.
	Logger *slog.Logger
}

// Server is the APEX HTTP API server.
// Server is safe for concurrent use after construction.
type Server struct {
	logger *slog.Logger
	http   *http.Server
}

// New validates cfg, applies defaults, wraps the handler with middleware,
// and returns a *Server ready to serve.
func New(cfg Config) (*Server, error) {
	if cfg.Handler == nil {
		return nil, errors.New("server: Handler must not be nil")
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	s := &Server{
		logger: cfg.Logger,
	}

	s.http = &http.Server{
		Addr:    cfg.Addr,
		Handler: middleware.CORS(cfg.Handler),
	}

	return s, nil
}

// Run starts the HTTP listener and blocks until ctx is cancelled.
// On cancellation it initiates a graceful shutdown, draining in-flight
// requests for up to shutdownTimeout before returning.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutCtx); err != nil {
		return err
	}
	return <-errCh
}
