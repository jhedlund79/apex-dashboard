// Package cache provides a TTL-based in-memory caching wrapper for
// [repository.Store]. Each endpoint result is cached independently.
//
// This is important for live data sources subject to API rate limits.
// A 60-second TTL keeps the dashboard well within the Polygon.io free
// tier limit of 5 requests per minute.
package cache

import (
	"sync"
	"time"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// entry holds a single cached result with its error and timestamp.
type entry[T any] struct {
	value T
	err   error
	at    time.Time
}

func (e *entry[T]) fresh(ttl time.Duration) bool {
	return !e.at.IsZero() && time.Since(e.at) < ttl
}

// Store wraps a [repository.Store] with per-method TTL caching.
// All methods are safe for concurrent use.
type Store struct {
	inner repository.Store
	ttl   time.Duration

	mu            sync.Mutex
	overview      entry[models.OverviewResponse]
	markets       entry[models.MarketsResponse]
	crypto        entry[models.CryptoResponse]
	recs          entry[models.RecommendationsResponse]
	principal     entry[models.PortfolioResponse]
	morganStanley entry[models.PortfolioResponse]
	fidelity      entry[models.PortfolioResponse]
	sofi          entry[models.PortfolioResponse]
}

// New wraps inner with a cache that expires entries after ttl.
func New(inner repository.Store, ttl time.Duration) *Store {
	return &Store{inner: inner, ttl: ttl}
}

func (s *Store) Overview() (models.OverviewResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.overview.fresh(s.ttl) {
		return s.overview.value, s.overview.err
	}
	v, err := s.inner.Overview()
	s.overview = entry[models.OverviewResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) Markets() (models.MarketsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.markets.fresh(s.ttl) {
		return s.markets.value, s.markets.err
	}
	v, err := s.inner.Markets()
	s.markets = entry[models.MarketsResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) Crypto() (models.CryptoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.crypto.fresh(s.ttl) {
		return s.crypto.value, s.crypto.err
	}
	v, err := s.inner.Crypto()
	s.crypto = entry[models.CryptoResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) Recommendations() (models.RecommendationsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.recs.fresh(s.ttl) {
		return s.recs.value, s.recs.err
	}
	v, err := s.inner.Recommendations()
	s.recs = entry[models.RecommendationsResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) Principal() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.principal.fresh(s.ttl) {
		return s.principal.value, s.principal.err
	}
	v, err := s.inner.Principal()
	s.principal = entry[models.PortfolioResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) MorganStanley() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.morganStanley.fresh(s.ttl) {
		return s.morganStanley.value, s.morganStanley.err
	}
	v, err := s.inner.MorganStanley()
	s.morganStanley = entry[models.PortfolioResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) Fidelity() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fidelity.fresh(s.ttl) {
		return s.fidelity.value, s.fidelity.err
	}
	v, err := s.inner.Fidelity()
	s.fidelity = entry[models.PortfolioResponse]{value: v, err: err, at: time.Now()}
	return v, err
}

func (s *Store) SoFi() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sofi.fresh(s.ttl) {
		return s.sofi.value, s.sofi.err
	}
	v, err := s.inner.SoFi()
	s.sofi = entry[models.PortfolioResponse]{value: v, err: err, at: time.Now()}
	return v, err
}
