package live

import (
	"fmt"

	"apex-dashboard/backend/internal/models"
)

// FetchQuote returns live price and change data for a single equity ticker.
// It reuses the same grouped-bars cache as all other methods so no extra API
// calls are made when the day's data has already been fetched.
func (s *Store) FetchQuote(ticker string) (models.QuoteResponse, error) {
	snaps, err := s.fetchSnapshots([]string{ticker})
	if err != nil {
		return models.QuoteResponse{}, fmt.Errorf("quote %s: %w", ticker, err)
	}
	snap, ok := snaps[ticker]
	if !ok {
		return models.QuoteResponse{}, fmt.Errorf("quote: ticker %q not found in market data", ticker)
	}
	dir := "neutral"
	if snap.TodaysChange > 0 {
		dir = "up"
	} else if snap.TodaysChange < 0 {
		dir = "down"
	}
	return models.QuoteResponse{
		Ticker:    ticker,
		Price:     snap.Day.C,
		Change:    snap.TodaysChange,
		ChangePct: snap.TodaysChangePerc,
		Dir:       dir,
	}, nil
}
