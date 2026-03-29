package cache

import (
	"errors"
	"sync"
	"testing"
	"time"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// countingStore is a fake Store that counts how many times each method is called.
type countingStore struct {
	mu sync.Mutex

	overviewCalls       int
	marketsCalls        int
	cryptoCalls         int
	recsCalls           int
	principalCalls      int
	morganStanleyCalls  int

	overviewResp       models.OverviewResponse
	overviewErr        error
	marketsResp        models.MarketsResponse
	marketsErr         error
	cryptoResp         models.CryptoResponse
	cryptoErr          error
	recsResp           models.RecommendationsResponse
	recsErr            error
	principalResp      models.PortfolioResponse
	principalErr       error
	morganStanleyResp  models.PortfolioResponse
	morganStanleyErr   error
}

func (s *countingStore) Overview() (models.OverviewResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overviewCalls++
	return s.overviewResp, s.overviewErr
}
func (s *countingStore) Markets() (models.MarketsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.marketsCalls++
	return s.marketsResp, s.marketsErr
}
func (s *countingStore) Crypto() (models.CryptoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cryptoCalls++
	return s.cryptoResp, s.cryptoErr
}
func (s *countingStore) Recommendations() (models.RecommendationsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recsCalls++
	return s.recsResp, s.recsErr
}
func (s *countingStore) Principal() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.principalCalls++
	return s.principalResp, s.principalErr
}
func (s *countingStore) MorganStanley() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.morganStanleyCalls++
	return s.morganStanleyResp, s.morganStanleyErr
}

func TestNew_ReturnsStore(t *testing.T) {
	inner := &countingStore{}
	s := New(inner, time.Minute)
	if s == nil {
		t.Fatal("expected non-nil Store")
	}
}

func TestCache_DelegatesOnFirstCall(t *testing.T) {
	inner := &countingStore{
		overviewResp: models.OverviewResponse{
			Header: models.HeaderInfo{FearGreed: "GREED: 80"},
		},
	}
	s := New(inner, time.Minute)

	resp, err := s.Overview()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Header.FearGreed != "GREED: 80" {
		t.Errorf("FearGreed = %q, want GREED: 80", resp.Header.FearGreed)
	}
	if inner.overviewCalls != 1 {
		t.Errorf("inner called %d times, want 1", inner.overviewCalls)
	}
}

func TestCache_ReturnsFromCacheOnSecondCall(t *testing.T) {
	inner := &countingStore{
		overviewResp: models.OverviewResponse{
			Header: models.HeaderInfo{MarketStatus: "Open"},
		},
	}
	s := New(inner, time.Minute)

	s.Overview() //nolint
	s.Overview() //nolint

	if inner.overviewCalls != 1 {
		t.Errorf("inner called %d times, want 1 (second should be cached)", inner.overviewCalls)
	}
}

func TestCache_RefreshesAfterTTL(t *testing.T) {
	inner := &countingStore{}
	s := New(inner, 10*time.Millisecond)

	s.Overview() //nolint
	time.Sleep(20 * time.Millisecond)
	s.Overview() //nolint

	if inner.overviewCalls != 2 {
		t.Errorf("inner called %d times, want 2 after TTL expiry", inner.overviewCalls)
	}
}

func TestCache_CachesError(t *testing.T) {
	inner := &countingStore{overviewErr: errors.New("api down")}
	s := New(inner, time.Minute)

	_, err1 := s.Overview()
	_, err2 := s.Overview()

	if err1 == nil || err2 == nil {
		t.Fatal("expected errors from both calls")
	}
	if inner.overviewCalls != 1 {
		t.Errorf("inner called %d times, want 1 (error should be cached)", inner.overviewCalls)
	}
}

func TestCache_ErrUnavailable_PassedThrough(t *testing.T) {
	inner := &countingStore{overviewErr: repository.ErrUnavailable}
	s := New(inner, time.Minute)

	_, err := s.Overview()
	if !errors.Is(err, repository.ErrUnavailable) {
		t.Errorf("err = %v, want ErrUnavailable", err)
	}
}

func TestCache_MethodsAreCachedIndependently(t *testing.T) {
	inner := &countingStore{}
	s := New(inner, time.Minute)

	s.Overview()         //nolint
	s.Markets()          //nolint
	s.Crypto()           //nolint
	s.Recommendations()  //nolint
	s.Principal()        //nolint
	s.MorganStanley()    //nolint

	// Call each a second time — all should hit cache.
	s.Overview()         //nolint
	s.Markets()          //nolint
	s.Crypto()           //nolint
	s.Recommendations()  //nolint
	s.Principal()        //nolint
	s.MorganStanley()    //nolint

	inner.mu.Lock()
	defer inner.mu.Unlock()
	checks := []struct {
		name  string
		calls int
	}{
		{"Overview", inner.overviewCalls},
		{"Markets", inner.marketsCalls},
		{"Crypto", inner.cryptoCalls},
		{"Recommendations", inner.recsCalls},
		{"Principal", inner.principalCalls},
		{"MorganStanley", inner.morganStanleyCalls},
	}
	for _, c := range checks {
		if c.calls != 1 {
			t.Errorf("%s: inner called %d times, want 1", c.name, c.calls)
		}
	}
}

func TestCache_AllMethodsReturnCorrectData(t *testing.T) {
	inner := &countingStore{
		marketsResp:       models.MarketsResponse{USSummary: models.USSummary{MarchReturn: "+5%"}},
		cryptoResp:        models.CryptoResponse{MarketContext: "bullish"},
		recsResp:          models.RecommendationsResponse{AISemis: []models.Recommendation{{Ticker: "NVDA"}}},
		principalResp:     models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$100k"}},
		morganStanleyResp: models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$50k"}},
	}
	s := New(inner, time.Minute)

	if r, _ := s.Markets(); r.USSummary.MarchReturn != "+5%" {
		t.Errorf("Markets: MarchReturn = %q, want +5%%", r.USSummary.MarchReturn)
	}
	if r, _ := s.Crypto(); r.MarketContext != "bullish" {
		t.Errorf("Crypto: MarketContext = %q, want bullish", r.MarketContext)
	}
	if r, _ := s.Recommendations(); len(r.AISemis) != 1 || r.AISemis[0].Ticker != "NVDA" {
		t.Errorf("Recommendations: unexpected aiSemis: %v", r.AISemis)
	}
	if r, _ := s.Principal(); r.Summary.CurrentValue != "$100k" {
		t.Errorf("Principal: CurrentValue = %q, want $100k", r.Summary.CurrentValue)
	}
	if r, _ := s.MorganStanley(); r.Summary.CurrentValue != "$50k" {
		t.Errorf("MorganStanley: CurrentValue = %q, want $50k", r.Summary.CurrentValue)
	}
}

func TestCache_ConcurrentCallsSafe(t *testing.T) {
	inner := &countingStore{
		overviewResp: models.OverviewResponse{Header: models.HeaderInfo{FearGreed: "50"}},
	}
	// Very short TTL so some goroutines refresh.
	s := New(inner, 5*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Overview() //nolint
		}()
	}
	wg.Wait()
	// No assertion needed — we're checking for races (run with -race).
}
