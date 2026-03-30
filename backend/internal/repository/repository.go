// Package repository defines the data-access boundary for the APEX dashboard.
//
// The [Store] interface is the only contract consumers depend on.
// Implementations must be safe for concurrent use.
//
// Use [static.New] to obtain a Store backed by hardcoded seed data, or
// provide your own implementation backed by a live market data API.
package repository

import (
	"errors"

	"apex-dashboard/backend/internal/models"
)

// ErrUnavailable is returned by a Store when it cannot produce data,
// for example due to a downstream timeout or missing configuration.
var ErrUnavailable = errors.New("repository: data unavailable")

// Store provides market data for all dashboard domains.
// Implementations must be safe for concurrent use.
type Store interface {
	// Overview returns aggregated data for the Overview page including
	// index performance, quick stats, chart series, and market drivers.
	Overview() (models.OverviewResponse, error)

	// Markets returns US market movers, international indices, and
	// sector performance data.
	Markets() (models.MarketsResponse, error)

	// Crypto returns crypto asset prices, small-cap entries, BTC price
	// history, and a market context narrative.
	Crypto() (models.CryptoResponse, error)

	// Recommendations returns curated growth picks grouped by theme.
	Recommendations() (models.RecommendationsResponse, error)

	// Principal returns portfolio data for the Principal.com retirement account.
	Principal() (models.PortfolioResponse, error)

	// MorganStanley returns portfolio data for the Morgan Stanley investment account.
	MorganStanley() (models.PortfolioResponse, error)

	// Fidelity returns portfolio data for the Fidelity brokerage account.
	Fidelity() (models.PortfolioResponse, error)

	// SoFi returns portfolio data for the SoFi Invest brokerage account.
	SoFi() (models.PortfolioResponse, error)
}
