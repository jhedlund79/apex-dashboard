// Package simulation provides a persistent store for the investment simulation feature.
// State is written atomically to a JSON file so the simulation survives server restarts.
package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"apex-dashboard/backend/internal/models"
)

const startingCash = 100_000.0

// Store persists simulation state to a JSON file.
// All methods are safe for concurrent use.
type Store struct {
	mu   sync.Mutex
	path string
	data models.SimulationAccount
}

// New loads or initialises a Store backed by the given file path.
// If the file does not exist it is created with the default starting balance.
func New(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if os.IsNotExist(err) {
		s.data = models.SimulationAccount{
			CashBalance:  startingCash,
			StartingCash: startingCash,
			Positions:    []models.SimulationPosition{},
			Transactions: []models.SimulationTx{},
			Snapshots:    []models.PortfolioSnapshot{},
		}
		return s.save()
	}
	if err != nil {
		return fmt.Errorf("simulation: open %s: %w", s.path, err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&s.data); err != nil {
		return fmt.Errorf("simulation: decode %s: %w", s.path, err)
	}
	// Ensure slices are never nil so JSON marshals as [] not null.
	if s.data.Positions == nil {
		s.data.Positions = []models.SimulationPosition{}
	}
	if s.data.Transactions == nil {
		s.data.Transactions = []models.SimulationTx{}
	}
	if s.data.Snapshots == nil {
		s.data.Snapshots = []models.PortfolioSnapshot{}
	}
	return nil
}

func (s *Store) save() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("simulation: create tmp: %w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s.data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("simulation: encode: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("simulation: close tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("simulation: rename: %w", err)
	}
	return nil
}

// Summary returns a copy of the current account state.
func (s *Store) Summary() (models.SimulationAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data, nil
}

// Trade records a buy or sell transaction and updates positions and cash.
func (s *Store) Trade(action, ticker string, shares, price float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := shares * price
	now := time.Now().Format("2006-01-02")

	switch action {
	case "buy":
		if total > s.data.CashBalance {
			return fmt.Errorf("insufficient funds: need $%.2f, have $%.2f", total, s.data.CashBalance)
		}
		s.data.CashBalance -= total
		s.updatePositionBuy(ticker, shares, price, now)
	case "sell":
		if err := s.updatePositionSell(ticker, shares); err != nil {
			return err
		}
		s.data.CashBalance += total
	default:
		return fmt.Errorf("invalid action %q: must be buy or sell", action)
	}

	tx := models.SimulationTx{
		ID:     fmt.Sprintf("txn_%d_%d", time.Now().UnixNano(), rand.Intn(10000)),
		Date:   now,
		Ticker: ticker,
		Action: action,
		Shares: shares,
		Price:  price,
		Total:  total,
	}
	// Prepend so most-recent appears first.
	s.data.Transactions = append([]models.SimulationTx{tx}, s.data.Transactions...)

	return s.save()
}

func (s *Store) updatePositionBuy(ticker string, shares, price float64, date string) {
	for i, p := range s.data.Positions {
		if p.Ticker == ticker {
			totalShares := p.Shares + shares
			s.data.Positions[i].AvgCost = (p.Shares*p.AvgCost + shares*price) / totalShares
			s.data.Positions[i].Shares = totalShares
			return
		}
	}
	s.data.Positions = append(s.data.Positions, models.SimulationPosition{
		Ticker:       ticker,
		Shares:       shares,
		AvgCost:      price,
		PurchaseDate: date,
	})
}

func (s *Store) updatePositionSell(ticker string, shares float64) error {
	for i, p := range s.data.Positions {
		if p.Ticker == ticker {
			if shares > p.Shares {
				return fmt.Errorf("insufficient shares: have %.4f %s, selling %.4f", p.Shares, ticker, shares)
			}
			if shares == p.Shares {
				s.data.Positions = append(s.data.Positions[:i], s.data.Positions[i+1:]...)
			} else {
				s.data.Positions[i].Shares -= shares
			}
			return nil
		}
	}
	return fmt.Errorf("no position found for %s", ticker)
}

// AddSnapshot appends (or updates) a portfolio value snapshot for today.
func (s *Store) AddSnapshot(totalValue float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	for i, snap := range s.data.Snapshots {
		if snap.Date == today {
			s.data.Snapshots[i].Value = totalValue
			return s.save()
		}
	}
	s.data.Snapshots = append(s.data.Snapshots, models.PortfolioSnapshot{
		Date:  today,
		Value: totalValue,
	})
	return s.save()
}

// Reset wipes all positions, transactions, and snapshots, restoring the starting cash.
func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = models.SimulationAccount{
		CashBalance:  startingCash,
		StartingCash: startingCash,
		Positions:    []models.SimulationPosition{},
		Transactions: []models.SimulationTx{},
		Snapshots:    []models.PortfolioSnapshot{},
	}
	return s.save()
}
