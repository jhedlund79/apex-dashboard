package live

import (
	"math"
	"testing"
	"time"

	"apex-dashboard/backend/internal/models"
)

// ── dirStr ────────────────────────────────────────────────────────────────────

func TestDirStr(t *testing.T) {
	cases := []struct {
		pct  float64
		want string
	}{
		{1.5, "up"},
		{0.001, "up"},
		{-1.5, "down"},
		{-0.001, "down"},
		{0, "neutral"},
	}
	for _, tc := range cases {
		got := dirStr(tc.pct)
		if got != tc.want {
			t.Errorf("dirStr(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

// ── fmtPct ────────────────────────────────────────────────────────────────────

func TestFmtPct(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{2.5, "+2.50%"},
		{0.0, "+0.00%"},
		{-1.23, "-1.23%"},
		{100.0, "+100.00%"},
	}
	for _, tc := range cases {
		got := fmtPct(tc.v)
		if got != tc.want {
			t.Errorf("fmtPct(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

// ── fmtDollar ────────────────────────────────────────────────────────────────

func TestFmtDollar(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{0.0001, "$0.0001"},
		{0.5, "$0.5000"},  // v < 1 → 4 decimal places
		{5.50, "$5.50"},
		{9.99, "$9.99"},
		{100, "$100.00"},  // 1 <= v < 1000 → 2 decimal places
		{999, "$999.00"},
		{1000, "$1,000"},  // v >= 1000 → no decimals, comma-separated
		{1234, "$1,234"},
		{12345, "$12,345"},
		{123456, "$123,456"},
		{1234567, "$1,234,567"},
	}
	for _, tc := range cases {
		got := fmtDollar(tc.v)
		if got != tc.want {
			t.Errorf("fmtDollar(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

// ── scoreFromPct ──────────────────────────────────────────────────────────────

func TestScoreFromPct(t *testing.T) {
	cases := []struct {
		pct, scale float64
		want       float64
	}{
		{0, 1, 50},       // neutral
		{50, 1, 100},     // clamped to max
		{-50, 1, 0},      // clamped to min
		{10, 2, 70},      // 50 + 10*2 = 70
		{-10, 2, 30},     // 50 + (-10)*2 = 30
		{1000, 1, 100},   // way over max → 100
		{-1000, 1, 0},    // way under min → 0
	}
	for _, tc := range cases {
		got := scoreFromPct(tc.pct, tc.scale)
		if got != tc.want {
			t.Errorf("scoreFromPct(%v, %v) = %v, want %v", tc.pct, tc.scale, got, tc.want)
		}
	}
}

// ── sampleBars ────────────────────────────────────────────────────────────────

func TestSampleBars_Empty(t *testing.T) {
	got := sampleBars(nil, 10)
	if len(got.Labels) != 0 || len(got.Data) != 0 {
		t.Errorf("expected empty ChartData for nil input, got %v", got)
	}
}

func TestSampleBars_FewerThanN(t *testing.T) {
	bars := []pgBar{
		{T: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 100},
		{T: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 200},
	}
	got := sampleBars(bars, 10)
	if len(got.Labels) != 2 {
		t.Errorf("labels = %d, want 2", len(got.Labels))
	}
	if got.Data[0] != 100 || got.Data[1] != 200 {
		t.Errorf("data = %v", got.Data)
	}
}

func TestSampleBars_MoreThanN(t *testing.T) {
	bars := make([]pgBar, 100)
	for i := range bars {
		bars[i] = pgBar{
			T: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * 24 * time.Hour).UnixMilli(),
			C: float64(i + 1),
		}
	}
	got := sampleBars(bars, 12)
	if len(got.Labels) > 12 {
		t.Errorf("labels = %d, want ≤12", len(got.Labels))
	}
}

func TestSampleBars_PriceRounded(t *testing.T) {
	bars := []pgBar{
		{T: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 123.456789},
	}
	got := sampleBars(bars, 1)
	want := math.Round(123.456789*100) / 100
	if got.Data[0] != want {
		t.Errorf("Data[0] = %v, want %v", got.Data[0], want)
	}
}

// ── samplePriceHistory ────────────────────────────────────────────────────────

func TestSamplePriceHistory_Empty(t *testing.T) {
	got := samplePriceHistory(nil, 10)
	if len(got.Labels) != 0 {
		t.Errorf("expected empty ChartData for nil input")
	}
}

func TestSamplePriceHistory_FewerThanN(t *testing.T) {
	prices := [][]float64{
		{float64(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()), 42000},
		{float64(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()), 43000},
	}
	got := samplePriceHistory(prices, 10)
	if len(got.Labels) != 2 {
		t.Errorf("labels = %d, want 2", len(got.Labels))
	}
}

func TestSamplePriceHistory_MoreThanN(t *testing.T) {
	prices := make([][]float64, 100)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range prices {
		prices[i] = []float64{float64(base.Add(time.Duration(i) * 24 * time.Hour).UnixMilli()), float64(i) * 100}
	}
	got := samplePriceHistory(prices, 12)
	if len(got.Labels) > 12 {
		t.Errorf("labels = %d, want ≤12", len(got.Labels))
	}
}

func TestSamplePriceHistory_PriceRounded(t *testing.T) {
	prices := [][]float64{
		{float64(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()), 42123.678},
	}
	got := samplePriceHistory(prices, 1)
	if got.Data[0] != math.Round(42123.678) {
		t.Errorf("Data[0] = %v, want %v", got.Data[0], math.Round(42123.678))
	}
}

// ── lastTradingDays ────────────────────────────────────────────────────────────

func TestLastTradingDays_FromWednesday(t *testing.T) {
	// Wednesday 2024-03-13
	now := time.Date(2024, 3, 13, 12, 0, 0, 0, time.UTC)
	day1, day2 := lastTradingDays(now)
	if day1 != "2024-03-12" {
		t.Errorf("day1 = %q, want 2024-03-12", day1)
	}
	if day2 != "2024-03-11" {
		t.Errorf("day2 = %q, want 2024-03-11", day2)
	}
}

func TestLastTradingDays_FromMonday(t *testing.T) {
	// Monday 2024-03-11 → prev trading days are Fri 2024-03-08, Thu 2024-03-07
	now := time.Date(2024, 3, 11, 12, 0, 0, 0, time.UTC)
	day1, day2 := lastTradingDays(now)
	if day1 != "2024-03-08" {
		t.Errorf("day1 = %q, want 2024-03-08 (Friday)", day1)
	}
	if day2 != "2024-03-07" {
		t.Errorf("day2 = %q, want 2024-03-07 (Thursday)", day2)
	}
}

func TestLastTradingDays_FromSunday(t *testing.T) {
	// Sunday 2024-03-10 → Fri 2024-03-08, Thu 2024-03-07
	now := time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC)
	day1, day2 := lastTradingDays(now)
	if day1 != "2024-03-08" {
		t.Errorf("day1 = %q, want 2024-03-08", day1)
	}
	if day2 != "2024-03-07" {
		t.Errorf("day2 = %q, want 2024-03-07", day2)
	}
}

func TestLastTradingDays_FromSaturday(t *testing.T) {
	// Saturday 2024-03-09 → Fri 2024-03-08, Thu 2024-03-07
	now := time.Date(2024, 3, 9, 12, 0, 0, 0, time.UTC)
	day1, day2 := lastTradingDays(now)
	if day1 != "2024-03-08" {
		t.Errorf("day1 = %q, want 2024-03-08", day1)
	}
	if day2 != "2024-03-07" {
		t.Errorf("day2 = %q, want 2024-03-07", day2)
	}
}

// ── marketStatus ──────────────────────────────────────────────────────────────

func TestMarketStatus_ReturnsNonEmptyString(t *testing.T) {
	status := marketStatus()
	if status == "" {
		t.Error("marketStatus() returned empty string")
	}
}

func TestMarketStatus_ContainsOpenOrClosed(t *testing.T) {
	status := marketStatus()
	if status != "Markets Open · "+time.Now().In(func() *time.Location {
		l, _ := time.LoadLocation("America/New_York")
		return l
	}()).Format("15:04")+" ET" && len(status) > 0 {
		// Just verify it starts with "Markets"
		if len(status) < 7 || status[:7] != "Markets" {
			t.Errorf("marketStatus() = %q, want string starting with 'Markets'", status)
		}
	}
}

// ── models.ChartData zero value ────────────────────────────────────────────────

func TestSampleBars_ReturnsChartDataType(t *testing.T) {
	bars := []pgBar{{T: time.Now().UnixMilli(), C: 500}}
	got := sampleBars(bars, 5)
	var _ models.ChartData = got // compile-time check
	if len(got.Labels) == 0 {
		t.Error("expected at least one label")
	}
}
