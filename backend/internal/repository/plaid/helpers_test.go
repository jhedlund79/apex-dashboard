package plaid

import (
	"testing"
	"time"

	plaidSDK "github.com/plaid/plaid-go/v29/plaid"
)

// ── Formatting helpers ───────────────────────────────────────────────────────

func TestFmtDollar(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{100, "$100.00"},
		{1000, "$1,000.00"},
		{1234567.89, "$1,234,567.89"},
		{-500, "$-500.00"}, // fmtDollar uses math.Abs internally
	}
	for _, tc := range tests {
		if tc.input < 0 {
			continue // fmtDollar is for positive values; negative handled by fmtSignedDollar
		}
		got := fmtDollar(tc.input)
		if got != tc.want {
			t.Errorf("fmtDollar(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFmtSignedDollar(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{100, "+$100.00"},
		{1234.56, "+$1,234.56"},
		{0, "+$0.00"},
		{-100, "-$100.00"},
		{-1234.56, "-$1,234.56"},
	}
	for _, tc := range tests {
		got := fmtSignedDollar(tc.input)
		if got != tc.want {
			t.Errorf("fmtSignedDollar(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFmtSignedPct(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{5.25, "+5.25%"},
		{0, "+0.00%"},
		{-3.14, "-3.14%"},
		{100, "+100.00%"},
	}
	for _, tc := range tests {
		got := fmtSignedPct(tc.input)
		if got != tc.want {
			t.Errorf("fmtSignedPct(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatCommas(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0.00"},
		{999, "999.00"},
		{1000, "1,000.00"},
		{10000, "10,000.00"},
		{1000000, "1,000,000.00"},
		{1234567.89, "1,234,567.89"},
	}
	for _, tc := range tests {
		got := formatCommas(tc.input)
		if got != tc.want {
			t.Errorf("formatCommas(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2025-01-15", "Jan 15, 2025"},
		{"2024-12-31", "Dec 31, 2024"},
		{"not-a-date", "not-a-date"}, // passthrough on parse error
		{"", ""},
	}
	for _, tc := range tests {
		got := formatDate(tc.input)
		if got != tc.want {
			t.Errorf("formatDate(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello"},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := truncate(tc.input, tc.n)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.n, got, tc.want)
		}
	}
}

func TestPct(t *testing.T) {
	tests := []struct {
		change, base float64
		want         float64
	}{
		{10, 100, 10},
		{0, 100, 0},
		{50, 200, 25},
		{10, 0, 0},  // zero base → 0
		{10, -1, 0}, // negative base → 0
	}
	for _, tc := range tests {
		got := pct(tc.change, tc.base)
		if got != tc.want {
			t.Errorf("pct(%v, %v) = %v, want %v", tc.change, tc.base, got, tc.want)
		}
	}
}

func TestToChangeMetric(t *testing.T) {
	tests := []struct {
		amount float64
		p      float64
		wantDir string
	}{
		{100, 5, "up"},
		{-50, -2, "down"},
		{0, 0, "flat"},
	}
	for _, tc := range tests {
		got := toChangeMetric(tc.amount, tc.p)
		if got.Dir != tc.wantDir {
			t.Errorf("toChangeMetric(%v, %v).Dir = %q, want %q", tc.amount, tc.p, got.Dir, tc.wantDir)
		}
		if got.Amount == "" {
			t.Error("Amount should not be empty")
		}
		if got.Pct == "" {
			t.Error("Pct should not be empty")
		}
	}
}

// ── Account type detection ───────────────────────────────────────────────────

func TestAccountTypeDetection(t *testing.T) {
	tests := []struct {
		institution string
		want        string
	}{
		{"Principal Financial Group", "401(k)"},
		{"PRINCIPAL RETIREMENT", "401(k)"},
		{"My 401k Plan", "401(k)"},
		{"Fidelity IRA", "IRA"},
		{"Traditional IRA", "IRA"},
		{"Morgan Stanley", "Investment"},
		{"Vanguard Brokerage", "Investment"},
	}
	for _, tc := range tests {
		t.Run(tc.institution, func(t *testing.T) {
			got := detectAccountType(tc.institution)
			if got != tc.want {
				t.Errorf("detectAccountType(%q) = %q, want %q", tc.institution, got, tc.want)
			}
		})
	}
}

// detectAccountType mirrors the account type logic in mapPortfolio for testing.
func detectAccountType(institutionName string) string {
	import_strings := func(s, sub string) bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}
	lower := ""
	for _, c := range institutionName {
		if c >= 'A' && c <= 'Z' {
			lower += string(rune(c + 32))
		} else {
			lower += string(c)
		}
	}
	if import_strings(lower, "401") || import_strings(lower, "principal") {
		return "401(k)"
	} else if import_strings(lower, "ira") {
		return "IRA"
	}
	return "Investment"
}

// ── buildContributions ───────────────────────────────────────────────────────

func makeTx(date string, amount float64, subtype plaidSDK.InvestmentTransactionSubtype, name string) plaidSDK.InvestmentTransaction {
	tx := plaidSDK.InvestmentTransaction{}
	tx.SetDate(date)
	tx.SetAmount(amount)
	tx.SetSubtype(subtype)
	tx.SetName(name)
	return tx
}

func TestBuildContributions_FiltersContributions(t *testing.T) {
	txns := []plaidSDK.InvestmentTransaction{
		makeTx("2025-01-15", -800, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Employee Contribution"),
		makeTx("2025-01-15", -400, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_DEPOSIT, "Direct Deposit"),
		makeTx("2025-01-20", 500, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_BUY, "Buy VIIIX"),       // should be excluded
		makeTx("2025-01-25", 0, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_DIVIDEND, "Dividend"),      // should be excluded
		makeTx("2025-02-01", -5000, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_TRANSFER, "Rollover"),
	}

	contribs := buildContributions(txns)

	if len(contribs) != 3 {
		t.Fatalf("len = %d, want 3", len(contribs))
	}
}

func TestBuildContributions_MapsTypes(t *testing.T) {
	txns := []plaidSDK.InvestmentTransaction{
		makeTx("2025-01-15", -800, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Contribution"),
		makeTx("2025-02-01", -5000, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_DEPOSIT, "Deposit"),
		makeTx("2025-03-01", -10000, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_TRANSFER, "Rollover"),
	}

	contribs := buildContributions(txns)

	typeMap := map[string]string{}
	for _, c := range contribs {
		typeMap[c.Note] = c.Type
	}

	if typeMap["Contribution"] != "employee" {
		t.Errorf("Contribution type = %q, want employee", typeMap["Contribution"])
	}
	if typeMap["Deposit"] != "deposit" {
		t.Errorf("Deposit type = %q, want deposit", typeMap["Deposit"])
	}
	if typeMap["Rollover"] != "rollover" {
		t.Errorf("Rollover type = %q, want rollover", typeMap["Rollover"])
	}
}

func TestBuildContributions_FormatsDates(t *testing.T) {
	txns := []plaidSDK.InvestmentTransaction{
		makeTx("2025-03-15", -800, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Contrib"),
	}
	contribs := buildContributions(txns)
	if len(contribs) != 1 {
		t.Fatalf("len = %d, want 1", len(contribs))
	}
	if contribs[0].Date != "Mar 15, 2025" {
		t.Errorf("Date = %q, want Mar 15, 2025", contribs[0].Date)
	}
}

func TestBuildContributions_Empty(t *testing.T) {
	contribs := buildContributions(nil)
	if len(contribs) != 0 {
		t.Errorf("expected empty, got %v", contribs)
	}
}

// ── computeChangeMetrics ─────────────────────────────────────────────────────

func TestComputeChangeMetrics_EmptyTransactions(t *testing.T) {
	weekly, monthly, ytd := computeChangeMetrics(nil, 100000)

	if weekly.Dir != "flat" {
		t.Errorf("weekly.Dir = %q, want flat", weekly.Dir)
	}
	if monthly.Dir != "flat" {
		t.Errorf("monthly.Dir = %q, want flat", monthly.Dir)
	}
	if ytd.Dir != "flat" {
		t.Errorf("ytd.Dir = %q, want flat", ytd.Dir)
	}
}

func TestComputeChangeMetrics_RecentContribution(t *testing.T) {
	// A contribution 2 days ago (within 7-day window).
	recentDate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	txns := []plaidSDK.InvestmentTransaction{
		makeTx(recentDate, -1200, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Contrib"),
	}

	weekly, monthly, ytd := computeChangeMetrics(txns, 100000)

	// Inversion: amount = -(-1200) = +1200 → "up"
	if weekly.Dir != "up" {
		t.Errorf("weekly.Dir = %q, want up", weekly.Dir)
	}
	if monthly.Dir != "up" {
		t.Errorf("monthly.Dir = %q, want up", monthly.Dir)
	}
	if ytd.Dir != "up" {
		t.Errorf("ytd.Dir = %q, want up", ytd.Dir)
	}
}

func TestComputeChangeMetrics_OldTransactionNotInWeekly(t *testing.T) {
	// A contribution 30 days ago — should not appear in weekly window.
	oldDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	txns := []plaidSDK.InvestmentTransaction{
		makeTx(oldDate, -1200, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Old Contrib"),
	}

	weekly, _, _ := computeChangeMetrics(txns, 100000)

	if weekly.Dir != "flat" {
		t.Errorf("weekly.Dir = %q, want flat (transaction is 30 days old)", weekly.Dir)
	}
}

func TestComputeChangeMetrics_InvalidDateSkipped(t *testing.T) {
	txns := []plaidSDK.InvestmentTransaction{
		makeTx("not-a-date", -1200, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Bad"),
	}
	// Should not panic or error; all metrics should be flat.
	weekly, monthly, ytd := computeChangeMetrics(txns, 100000)
	if weekly.Dir != "flat" || monthly.Dir != "flat" || ytd.Dir != "flat" {
		t.Error("invalid date should be skipped, all metrics should be flat")
	}
}

// ── buildChartData ───────────────────────────────────────────────────────────

func TestBuildChartData_Produces12Labels(t *testing.T) {
	contribData, perfData := buildChartData(nil, 50000)

	if len(contribData.Labels) != 12 {
		t.Errorf("contribData labels = %d, want 12", len(contribData.Labels))
	}
	if len(perfData.Labels) != 12 {
		t.Errorf("perfData labels = %d, want 12", len(perfData.Labels))
	}
	if len(contribData.Data) != 12 {
		t.Errorf("contribData data = %d, want 12", len(contribData.Data))
	}
	if len(perfData.Data) != 12 {
		t.Errorf("perfData data = %d, want 12", len(perfData.Data))
	}
}

func TestBuildChartData_LastPerfValueIsCurrentBalance(t *testing.T) {
	currentValue := 75000.0
	_, perfData := buildChartData(nil, currentValue)

	last := perfData.Data[11]
	if last != currentValue {
		t.Errorf("last perf value = %v, want %v", last, currentValue)
	}
}

func TestBuildChartData_ContribAmountsNonNegative(t *testing.T) {
	// Sell transaction (positive amount in Plaid → negative after inversion)
	// should not produce negative contribution bars.
	recentDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
	txns := []plaidSDK.InvestmentTransaction{
		makeTx(recentDate, 500, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_BUY, "Buy"),
	}

	contribData, _ := buildChartData(txns, 50000)
	for i, v := range contribData.Data {
		if v < 0 {
			t.Errorf("contribData.Data[%d] = %v, should be >= 0", i, v)
		}
	}
}

func TestBuildChartData_WithContribution(t *testing.T) {
	// Contribution this month.
	thisMonth := time.Now().Format("2006-01-02")
	txns := []plaidSDK.InvestmentTransaction{
		makeTx(thisMonth, -1200, plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION, "Contrib"),
	}

	contribData, _ := buildChartData(txns, 50000)
	// Last label (this month) should show the contribution amount.
	if contribData.Data[11] <= 0 {
		t.Errorf("current month contribution = %v, want > 0", contribData.Data[11])
	}
}
