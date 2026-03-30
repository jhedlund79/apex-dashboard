package simulation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"apex-dashboard/backend/internal/models"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "sim.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

// ── New / load ────────────────────────────────────────────────────────────────

func TestNew_CreatesFileWithDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// File should exist.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}

	acct, _ := s.Summary()
	if acct.CashBalance != startingCash {
		t.Errorf("CashBalance = %v, want %v", acct.CashBalance, startingCash)
	}
	if acct.StartingCash != startingCash {
		t.Errorf("StartingCash = %v, want %v", acct.StartingCash, startingCash)
	}
	if acct.Positions == nil {
		t.Error("Positions should be non-nil slice")
	}
	if acct.Transactions == nil {
		t.Error("Transactions should be non-nil slice")
	}
	if acct.Snapshots == nil {
		t.Error("Snapshots should be non-nil slice")
	}
}

func TestNew_LoadsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	// Write a pre-existing state.
	existing := models.SimulationAccount{
		CashBalance:  75_000,
		StartingCash: 100_000,
		Positions:    []models.SimulationPosition{{Ticker: "AAPL", Shares: 10, AvgCost: 150}},
		Transactions: []models.SimulationTx{},
		Snapshots:    []models.PortfolioSnapshot{},
	}
	f, _ := os.Create(path)
	json.NewEncoder(f).Encode(existing) //nolint
	f.Close()

	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	acct, _ := s.Summary()
	if acct.CashBalance != 75_000 {
		t.Errorf("CashBalance = %v, want 75000", acct.CashBalance)
	}
	if len(acct.Positions) != 1 || acct.Positions[0].Ticker != "AAPL" {
		t.Errorf("Positions = %v", acct.Positions)
	}
}

func TestNew_CorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")
	os.WriteFile(path, []byte("not valid json{{{"), 0644) //nolint

	_, err := New(path)
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}

func TestNew_NilSlicesNormalised(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	// Write JSON with explicit null slices.
	os.WriteFile(path, []byte(`{"cashBalance":100000,"startingCash":100000,"positions":null,"transactions":null,"snapshots":null}`), 0644) //nolint

	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	acct, _ := s.Summary()
	if acct.Positions == nil {
		t.Error("Positions should be non-nil after load")
	}
	if acct.Transactions == nil {
		t.Error("Transactions should be non-nil after load")
	}
	if acct.Snapshots == nil {
		t.Error("Snapshots should be non-nil after load")
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSummary_ReturnsCopy(t *testing.T) {
	s := newTestStore(t)
	a1, _ := s.Summary()
	a2, _ := s.Summary()
	if a1.CashBalance != a2.CashBalance {
		t.Error("expected same cash balance on repeated calls")
	}
}

// ── Trade – buy ───────────────────────────────────────────────────────────────

func TestTrade_BuyNewPosition(t *testing.T) {
	s := newTestStore(t)

	if err := s.Trade("buy", "NVDA", 10, 100); err != nil {
		t.Fatalf("Trade: %v", err)
	}

	acct, _ := s.Summary()
	if acct.CashBalance != startingCash-1000 {
		t.Errorf("CashBalance = %v, want %v", acct.CashBalance, startingCash-1000)
	}
	if len(acct.Positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(acct.Positions))
	}
	if acct.Positions[0].Ticker != "NVDA" {
		t.Errorf("Ticker = %q, want NVDA", acct.Positions[0].Ticker)
	}
	if acct.Positions[0].Shares != 10 {
		t.Errorf("Shares = %v, want 10", acct.Positions[0].Shares)
	}
	if acct.Positions[0].AvgCost != 100 {
		t.Errorf("AvgCost = %v, want 100", acct.Positions[0].AvgCost)
	}
	if len(acct.Transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(acct.Transactions))
	}
	if acct.Transactions[0].Action != "buy" {
		t.Errorf("Action = %q, want buy", acct.Transactions[0].Action)
	}
	if acct.Transactions[0].Total != 1000 {
		t.Errorf("Total = %v, want 1000", acct.Transactions[0].Total)
	}
}

func TestTrade_BuyUpdatesAverageCost(t *testing.T) {
	s := newTestStore(t)

	s.Trade("buy", "NVDA", 10, 100) //nolint
	s.Trade("buy", "NVDA", 10, 200) //nolint

	acct, _ := s.Summary()
	if len(acct.Positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(acct.Positions))
	}
	pos := acct.Positions[0]
	if pos.Shares != 20 {
		t.Errorf("Shares = %v, want 20", pos.Shares)
	}
	// avg = (10*100 + 10*200) / 20 = 150
	if pos.AvgCost != 150 {
		t.Errorf("AvgCost = %v, want 150", pos.AvgCost)
	}
}

func TestTrade_BuyInsufficientFunds(t *testing.T) {
	s := newTestStore(t)
	// Costs more than starting cash.
	err := s.Trade("buy", "BRK", 10, startingCash)
	if err == nil {
		t.Error("expected error for insufficient funds")
	}
}

func TestTrade_BuyMultipleTickers(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 5, 100) //nolint
	s.Trade("buy", "AAPL", 5, 200) //nolint

	acct, _ := s.Summary()
	if len(acct.Positions) != 2 {
		t.Fatalf("positions = %d, want 2", len(acct.Positions))
	}
}

// ── Trade – sell ──────────────────────────────────────────────────────────────

func TestTrade_SellFullPosition(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 10, 100) //nolint

	if err := s.Trade("sell", "NVDA", 10, 150); err != nil {
		t.Fatalf("Trade sell: %v", err)
	}

	acct, _ := s.Summary()
	if len(acct.Positions) != 0 {
		t.Errorf("positions = %d, want 0 after full sell", len(acct.Positions))
	}
	// Cash: 100000 - 1000 + 1500 = 100500
	if acct.CashBalance != startingCash-1000+1500 {
		t.Errorf("CashBalance = %v, want %v", acct.CashBalance, startingCash-1000+1500)
	}
}

func TestTrade_SellPartialPosition(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 10, 100) //nolint

	if err := s.Trade("sell", "NVDA", 5, 120); err != nil {
		t.Fatalf("Trade partial sell: %v", err)
	}

	acct, _ := s.Summary()
	if len(acct.Positions) != 1 {
		t.Fatalf("positions = %d, want 1 after partial sell", len(acct.Positions))
	}
	if acct.Positions[0].Shares != 5 {
		t.Errorf("Shares = %v, want 5", acct.Positions[0].Shares)
	}
}

func TestTrade_SellMoreThanOwned(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 10, 100) //nolint

	err := s.Trade("sell", "NVDA", 20, 100)
	if err == nil {
		t.Error("expected error when selling more than owned")
	}
}

func TestTrade_SellNonExistentPosition(t *testing.T) {
	s := newTestStore(t)
	err := s.Trade("sell", "NVDA", 1, 100)
	if err == nil {
		t.Error("expected error when selling non-existent position")
	}
}

// ── Trade – invalid action ─────────────────────────────────────────────────────

func TestTrade_InvalidAction(t *testing.T) {
	s := newTestStore(t)
	err := s.Trade("hold", "NVDA", 1, 100)
	if err == nil {
		t.Error("expected error for invalid action")
	}
}

// ── Trade – transaction ordering ─────────────────────────────────────────────

func TestTrade_MostRecentTransactionFirst(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 1, 100)  //nolint
	s.Trade("buy", "AAPL", 1, 200)  //nolint

	acct, _ := s.Summary()
	// Most recent buy (AAPL) should be first.
	if acct.Transactions[0].Ticker != "AAPL" {
		t.Errorf("Transactions[0].Ticker = %q, want AAPL", acct.Transactions[0].Ticker)
	}
}

// ── Trade – persistence ────────────────────────────────────────────────────────

func TestTrade_PersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	s1, _ := New(path)
	s1.Trade("buy", "NVDA", 5, 100) //nolint

	// Open a new Store from the same file.
	s2, err := New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	acct, _ := s2.Summary()
	if len(acct.Positions) != 1 || acct.Positions[0].Ticker != "NVDA" {
		t.Errorf("position not persisted: %v", acct.Positions)
	}
}

// ── AddSnapshot ───────────────────────────────────────────────────────────────

func TestAddSnapshot_AddsEntry(t *testing.T) {
	s := newTestStore(t)
	if err := s.AddSnapshot(105_000); err != nil {
		t.Fatalf("AddSnapshot: %v", err)
	}
	acct, _ := s.Summary()
	if len(acct.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1", len(acct.Snapshots))
	}
	if acct.Snapshots[0].Value != 105_000 {
		t.Errorf("Value = %v, want 105000", acct.Snapshots[0].Value)
	}
}

func TestAddSnapshot_UpdatesSameDayEntry(t *testing.T) {
	s := newTestStore(t)
	s.AddSnapshot(105_000) //nolint
	s.AddSnapshot(110_000) //nolint

	acct, _ := s.Summary()
	// Same day → only one entry, updated value.
	if len(acct.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1 (same day update)", len(acct.Snapshots))
	}
	if acct.Snapshots[0].Value != 110_000 {
		t.Errorf("Value = %v, want 110000", acct.Snapshots[0].Value)
	}
}

// ── Reset ─────────────────────────────────────────────────────────────────────

func TestReset_ClearsEverything(t *testing.T) {
	s := newTestStore(t)
	s.Trade("buy", "NVDA", 5, 100) //nolint
	s.AddSnapshot(99_000)          //nolint

	if err := s.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	acct, _ := s.Summary()
	if acct.CashBalance != startingCash {
		t.Errorf("CashBalance = %v, want %v", acct.CashBalance, startingCash)
	}
	if len(acct.Positions) != 0 {
		t.Errorf("positions = %v, want empty", acct.Positions)
	}
	if len(acct.Transactions) != 0 {
		t.Errorf("transactions = %v, want empty", acct.Transactions)
	}
	if len(acct.Snapshots) != 0 {
		t.Errorf("snapshots = %v, want empty", acct.Snapshots)
	}
}

func TestReset_PersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	s1, _ := New(path)
	s1.Trade("buy", "NVDA", 5, 100) //nolint
	s1.Reset()                       //nolint

	s2, _ := New(path)
	acct, _ := s2.Summary()
	if len(acct.Positions) != 0 {
		t.Errorf("positions after reset reload = %v, want empty", acct.Positions)
	}
	if acct.CashBalance != startingCash {
		t.Errorf("CashBalance = %v, want %v", acct.CashBalance, startingCash)
	}
}

// ── save error path ───────────────────────────────────────────────────────────

func TestTrade_SaveFailsOnReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sim.json")

	// Create a valid store first.
	s, _ := New(path)

	// Make directory read-only so tmp file can't be created.
	os.Chmod(dir, 0555) //nolint
	defer os.Chmod(dir, 0755)

	err := s.Trade("buy", "NVDA", 1, 100)
	if err == nil {
		t.Error("expected save error on read-only dir")
	}
}
