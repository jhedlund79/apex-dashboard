// Package config loads runtime configuration from the environment.
// Call [Load] once at startup and pass the resulting [Config] to
// dependent packages; avoid calling os.Getenv outside this package.
package config

import (
	"bufio"
	"os"
	"strings"
)

// Config holds all runtime configuration for the server.
type Config struct {
	// Addr is the TCP address to listen on (e.g. ":8080").
	Addr string

	// PolygonKey is the Polygon.io API key used by the live data store.
	// When empty the server falls back to the static seed-data store.
	PolygonKey string

	// CoinGeckoKey is an optional CoinGecko API key.
	// The public API works without one but is subject to stricter rate limits.
	CoinGeckoKey string

	// Plaid credentials. All three must be set to enable Plaid integration.
	PlaidClientID string
	PlaidSecret   string
	// PlaidEnv controls which Plaid environment to use: "sandbox", "development",
	// or "production". Defaults to "sandbox" when unset.
	PlaidEnv string
}

// Load reads configuration from environment variables and returns a
// populated Config. Unset variables fall back to documented defaults.
//
// Load sources a .env file if one is found. It checks, in order:
//  1. .env in the current working directory
//  2. .env in the parent directory (useful when running from a sub-directory)
//
// Explicit environment variables always take precedence over .env values.
//
// Environment variables:
//
//	APEX_ADDR          — listen address (default ":8080")
//	PORT               — Cloud Run port (takes precedence over APEX_ADDR)
//	POLYGON_API_KEY    — Polygon.io API key; enables live market data
//	COINGECKO_API_KEY  — optional CoinGecko API key
func Load() Config {
	loadDotEnv(".env", "../.env")

	addr := os.Getenv("APEX_ADDR")
	if addr == "" {
		if port := os.Getenv("PORT"); port != "" {
			addr = ":" + port
		} else {
			addr = ":8080"
		}
	}
	plaidEnv := os.Getenv("PLAID_ENV")
	if plaidEnv == "" {
		plaidEnv = "sandbox"
	}
	return Config{
		Addr:          addr,
		PolygonKey:    os.Getenv("POLYGON_API_KEY"),
		CoinGeckoKey:  os.Getenv("COINGECKO_API_KEY"),
		PlaidClientID: os.Getenv("PLAID_CLIENT_ID"),
		PlaidSecret:   os.Getenv("PLAID_SECRET"),
		PlaidEnv:      plaidEnv,
	}
}

// loadDotEnv reads key=value pairs from the first path that exists and sets
// any that are not already present in the environment.
// Lines beginning with # and blank lines are ignored.
func loadDotEnv(paths ...string) {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			// Strip optional surrounding quotes
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			// Only set if not already defined — explicit env vars win
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
		f.Close()
		return // stop at first file found
	}
}
