package live

import (
	"testing"
)

// ── rankByMomentum ────────────────────────────────────────────────────────────

func TestRankByMomentum_SelectsTopN(t *testing.T) {
	s := newTestStore(nil)
	candidates := []candidate{
		{ticker: "A", name: "Alpha"},
		{ticker: "B", name: "Beta"},
		{ticker: "C", name: "Gamma"},
		{ticker: "D", name: "Delta"},
	}
	todayBars := map[string]float64{"A": 110, "B": 105, "C": 120, "D": 108}
	oneMonthBars := map[string]float64{"A": 100, "B": 100, "C": 100, "D": 100}

	result := s.rankByMomentum(candidates, todayBars, oneMonthBars, 2)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	// C has highest momentum (+20%), should be first
	if result[0].ticker != "C" {
		t.Errorf("first = %q, want C", result[0].ticker)
	}
}

func TestRankByMomentum_ExcludesMissingData(t *testing.T) {
	s := newTestStore(nil)
	candidates := []candidate{
		{ticker: "A", name: "Alpha"},
		{ticker: "MISSING", name: "No Data"},
	}
	todayBars := map[string]float64{"A": 110}
	oneMonthBars := map[string]float64{"A": 100} // MISSING not present

	result := s.rankByMomentum(candidates, todayBars, oneMonthBars, 3)
	if len(result) != 1 {
		t.Errorf("len = %d, want 1 (MISSING excluded)", len(result))
	}
}

func TestRankByMomentum_ExcludesZeroPrice(t *testing.T) {
	s := newTestStore(nil)
	candidates := []candidate{
		{ticker: "A", name: "Alpha"},
		{ticker: "ZERO", name: "Zero Price"},
	}
	todayBars := map[string]float64{"A": 110, "ZERO": 0}
	oneMonthBars := map[string]float64{"A": 100, "ZERO": 90}

	result := s.rankByMomentum(candidates, todayBars, oneMonthBars, 3)
	if len(result) != 1 {
		t.Errorf("len = %d, want 1 (zero price excluded)", len(result))
	}
}

func TestRankByMomentum_FewerCandidatesThanN(t *testing.T) {
	s := newTestStore(nil)
	candidates := []candidate{
		{ticker: "A", name: "Alpha"},
	}
	todayBars := map[string]float64{"A": 110}
	oneMonthBars := map[string]float64{"A": 100}

	result := s.rankByMomentum(candidates, todayBars, oneMonthBars, 3)
	if len(result) != 1 {
		t.Errorf("len = %d, want 1", len(result))
	}
}

func TestRankByMomentum_SortsDescending(t *testing.T) {
	s := newTestStore(nil)
	candidates := []candidate{
		{ticker: "LOW", name: "Low Mom"},
		{ticker: "HIGH", name: "High Mom"},
		{ticker: "MID", name: "Mid Mom"},
	}
	todayBars := map[string]float64{"LOW": 101, "HIGH": 130, "MID": 115}
	oneMonthBars := map[string]float64{"LOW": 100, "HIGH": 100, "MID": 100}

	result := s.rankByMomentum(candidates, todayBars, oneMonthBars, 3)
	if result[0].ticker != "HIGH" || result[1].ticker != "MID" || result[2].ticker != "LOW" {
		t.Errorf("wrong order: %v %v %v", result[0].ticker, result[1].ticker, result[2].ticker)
	}
}

// ── rankCrypto ────────────────────────────────────────────────────────────────

func TestRankCrypto_SelectsTopN(t *testing.T) {
	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", CurrentPrice: 82000, PriceChangePerc24h: -2.0},
		{ID: "ethereum", Name: "Ethereum", CurrentPrice: 3000, PriceChangePerc24h: 3.5},
		{ID: "solana", Name: "Solana", CurrentPrice: 140, PriceChangePerc24h: 1.0},
	}
	candidates := []cgCandidate{
		{ticker: "BTC", cgID: "bitcoin", thesis: "T", badge: "", recType: "buy"},
		{ticker: "ETH", cgID: "ethereum", thesis: "T", badge: "", recType: "buy"},
		{ticker: "SOL", cgID: "solana", thesis: "T", badge: "", recType: "speculative"},
	}

	result := rankCrypto(coins, candidates, 2)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].ticker != "ETH" {
		t.Errorf("first = %q, want ETH (highest 24h change)", result[0].ticker)
	}
}

func TestRankCrypto_MissingCoinSkipped(t *testing.T) {
	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", CurrentPrice: 82000, PriceChangePerc24h: 1.0},
	}
	candidates := []cgCandidate{
		{ticker: "BTC", cgID: "bitcoin"},
		{ticker: "MISSING", cgID: "not-found"},
	}

	result := rankCrypto(coins, candidates, 3)
	if len(result) != 1 {
		t.Errorf("len = %d, want 1 (missing coin excluded)", len(result))
	}
}

func TestRankCrypto_Empty(t *testing.T) {
	result := rankCrypto(nil, nil, 3)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

// ── badgeForRank ──────────────────────────────────────────────────────────────

func TestBadgeForRank_Override(t *testing.T) {
	badge, recType := badgeForRank(0, "Custom Badge", "buy")
	if badge != "Custom Badge" || recType != "buy" {
		t.Errorf("badge=%q recType=%q", badge, recType)
	}
}

func TestBadgeForRank_Rank0(t *testing.T) {
	badge, recType := badgeForRank(0, "", "")
	if badge != "Strong Buy" || recType != "strong-buy" {
		t.Errorf("badge=%q recType=%q", badge, recType)
	}
}

func TestBadgeForRank_Rank1(t *testing.T) {
	badge, recType := badgeForRank(1, "", "")
	if badge != "Buy" || recType != "buy" {
		t.Errorf("badge=%q recType=%q", badge, recType)
	}
}

func TestBadgeForRank_Rank2Plus(t *testing.T) {
	badge, recType := badgeForRank(2, "", "")
	if badge != "Speculative" || recType != "speculative" {
		t.Errorf("badge=%q recType=%q", badge, recType)
	}
	badge, recType = badgeForRank(5, "", "")
	if badge != "Speculative" || recType != "speculative" {
		t.Errorf("badge=%q recType=%q", badge, recType)
	}
}

// ── fmtMom ────────────────────────────────────────────────────────────────────

func TestFmtMom(t *testing.T) {
	cases := []struct {
		pct  float64
		want string
	}{
		{10.5, "+10.5%"},
		{-5.2, "-5.2%"},
		{0, "+0.0%"},
	}
	for _, tc := range cases {
		got := fmtMom(tc.pct)
		if got != tc.want {
			t.Errorf("fmtMom(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}
