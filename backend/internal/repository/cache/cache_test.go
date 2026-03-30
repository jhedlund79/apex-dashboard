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

	overviewCalls      int
	marketsCalls       int
	cryptoCalls        int
	recsCalls          int
	principalCalls     int
	morganStanleyCalls int
	fidelityCalls      int
	sofiCalls          int

	overviewResp      models.OverviewResponse
	overviewErr       error
	marketsResp       models.MarketsResponse
	marketsErr        error
	cryptoResp        models.CryptoResponse
	cryptoErr         error
	recsResp          models.RecommendationsResponse
	recsErr           error
	principalResp     models.PortfolioResponse
	principalErr      error
	morganStanleyResp models.PortfolioResponse
	morganStanleyErr  error
	fidelityResp      models.PortfolioResponse
	fidelityErr       error
	sofiResp          models.PortfolioResponse
	sofiErr           error
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
func (s *countingStore) Fidelity() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fidelityCalls++
	return s.fidelityResp, s.fidelityErr
}
func (s *countingStore) SoFi() (models.PortfolioResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sofiCalls++
	return s.sofiResp, s.sofiErr
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

func TestCache_FidelityAndSoFi(t *testing.T) {
	inner := &countingStore{
		fidelityResp: models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$43k"}},
		sofiResp:     models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$8k"}},
	}
	s := New(inner, time.Minute)

	if r, _ := s.Fidelity(); r.Summary.CurrentValue != "$43k" {
		t.Errorf("Fidelity: CurrentValue = %q, want $43k", r.Summary.CurrentValue)
	}
	if r, _ := s.SoFi(); r.Summary.CurrentValue != "$8k" {
		t.Errorf("SoFi: CurrentValue = %q, want $8k", r.Summary.CurrentValue)
	}

	// Second call should hit cache.
	s.Fidelity() //nolint
	s.SoFi()     //nolint

	inner.mu.Lock()
	defer inner.mu.Unlock()
	if inner.fidelityCalls != 1 {
		t.Errorf("Fidelity inner called %d times, want 1", inner.fidelityCalls)
	}
	if inner.sofiCalls != 1 {
		t.Errorf("SoFi inner called %d times, want 1", inner.sofiCalls)
	}
}

func TestCache_FidelityAndSoFiIncludedInMethodsAreCachedIndependently(t *testing.T) {
	inner := &countingStore{}
	s := New(inner, time.Minute)

	s.Fidelity() //nolint
	s.SoFi()     //nolint
	s.Fidelity() //nolint
	s.SoFi()     //nolint

	inner.mu.Lock()
	defer inner.mu.Unlock()
	if inner.fidelityCalls != 1 {
		t.Errorf("Fidelity called %d times, want 1", inner.fidelityCalls)
	}
	if inner.sofiCalls != 1 {
		t.Errorf("SoFi called %d times, want 1", inner.sofiCalls)
	}
}

func TestCache_FidelityCachesError(t *testing.T) {
	inner := &countingStore{fidelityErr: errors.New("fidelity down")}
	s := New(inner, time.Minute)

	_, err1 := s.Fidelity()
	_, err2 := s.Fidelity()
	if err1 == nil || err2 == nil {
		t.Fatal("expected errors from both calls")
	}
	inner.mu.Lock()
	defer inner.mu.Unlock()
	if inner.fidelityCalls != 1 {
		t.Errorf("inner called %d times, want 1 (error should be cached)", inner.fidelityCalls)
	}
}

func TestCache_AllMethodsIncludingFidelitySoFi(t *testing.T) {
	inner := &countingStore{
		marketsResp:       models.MarketsResponse{USSummary: models.USSummary{MarchReturn: "+5%"}},
		cryptoResp:        models.CryptoResponse{MarketContext: "bullish"},
		recsResp:          models.RecommendationsResponse{AISemis: []models.Recommendation{{Ticker: "NVDA"}}},
		principalResp:     models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$100k"}},
		morganStanleyResp: models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$50k"}},
		fidelityResp:      models.PortfolioResponse{Connected: true, Summary: models.PortfolioSummary{CurrentValue: "$43k"}},
		sofiResp:          models.PortfolioResponse{Connected: false},
	}
	s := New(inner, time.Minute)

	// Call each twice; only first should hit inner.
	for i := 0; i < 2; i++ {
		s.Markets()       //nolint
		s.Crypto()        //nolint
		s.Recommendations() //nolint
		s.Principal()     //nolint
		s.MorganStanley() //nolint
		s.Fidelity()      //nolint
		s.SoFi()          //nolint
	}

	inner.mu.Lock()
	defer inner.mu.Unlock()
	for _, c := range []struct {
		name  string
		calls int
	}{
		{"Markets", inner.marketsCalls},
		{"Crypto", inner.cryptoCalls},
		{"Recommendations", inner.recsCalls},
		{"Principal", inner.principalCalls},
		{"MorganStanley", inner.morganStanleyCalls},
		{"Fidelity", inner.fidelityCalls},
		{"SoFi", inner.sofiCalls},
	} {
		if c.calls != 1 {
			t.Errorf("%s: inner called %d times, want 1", c.name, c.calls)
		}
	}
}
