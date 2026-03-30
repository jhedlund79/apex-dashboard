// Package handlers implements the HTTP handler layer for the APEX dashboard API.
//
// Create a Handler with [New], then register its route methods on an http.ServeMux.
// All handlers follow the standard response envelope:
//
//	Success: {"data": <payload>}
//	Error:   {"error": {"code": "<CODE>", "message": "<text>"}}
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// PlaidManager is the subset of the Plaid store used by the handler layer.
type PlaidManager interface {
	CreateLinkToken(ctx context.Context, slot string) (string, error)
	ExchangeToken(ctx context.Context, slot, publicToken string) error
	ConnectedSlots() []string
}

// Quoter provides live price data for simulation trades.
type Quoter interface {
	FetchQuote(ticker string) (models.QuoteResponse, error)
}

// SimulationManager manages the persistent simulation portfolio.
type SimulationManager interface {
	Summary() (models.SimulationAccount, error)
	Trade(action, ticker string, shares, price float64) error
	AddSnapshot(totalValue float64) error
	Reset() error
}

// ErrInvalidConfig is returned by [New] when required Config fields are missing.
var ErrInvalidConfig = errors.New("handlers: invalid config")

// Config controls Handler construction.
type Config struct {
	// Required: market data provider.
	Store repository.Store

	// Optional: enables /api/v1/plaid/* routes when set.
	Plaid PlaidManager

	// Optional: nil disables request logging.
	Logger *slog.Logger

	// Optional: enables /api/v1/simulation/* routes when set.
	Simulation SimulationManager

	// Optional: provides live quotes for simulation trades.
	Quoter Quoter
}

// Handler holds shared dependencies for all route handlers.
// It is safe for concurrent use after construction.
type Handler struct {
	store  repository.Store
	plaid  PlaidManager // may be nil if Plaid is not configured
	logger *slog.Logger
	sim    SimulationManager // may be nil if simulation is not configured
	quoter Quoter            // may be nil if live data is not available
}

// New validates cfg and returns a *Handler ready to register routes.
// Returns [ErrInvalidConfig] if cfg.Store is nil.
func New(cfg Config) (*Handler, error) {
	if cfg.Store == nil {
		return nil, ErrInvalidConfig
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Handler{
		store:  cfg.Store,
		plaid:  cfg.Plaid,
		logger: cfg.Logger,
		sim:    cfg.Simulation,
		quoter: cfg.Quoter,
	}, nil
}

// respond writes a standard {"data": data} envelope as JSON with the given
// HTTP status code. On encode failure it logs the error and falls back to a
// 500 plain-text response.
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]any{"data": data}); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to encode response", "err", err)
	}
}

// storeError writes an appropriate HTTP error response for a Store failure.
// [repository.ErrUnavailable] maps to 503; all other errors map to 500.
func (h *Handler) storeError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "store error", "err", err)

	type apiError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	type envelope struct {
		Error apiError `json:"error"`
	}

	w.Header().Set("Content-Type", "application/json")

	if errors.Is(err, repository.ErrUnavailable) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(envelope{Error: apiError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "market data unavailable",
		}})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(envelope{Error: apiError{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	}})
}
