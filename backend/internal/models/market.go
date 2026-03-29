// Package models defines all JSON response types for the APEX dashboard API.
// Types are plain structs with no interfaces; they are used by both the
// repository layer (as return values) and the handler layer (as response bodies).
package models

// HeaderInfo carries the live status line and fear/greed index for the
// dashboard header.
type HeaderInfo struct {
	FearGreed    string `json:"fearGreed"`
	MarketStatus string `json:"marketStatus"`
}

// Index represents a single market index card.
type Index struct {
	Ticker    string `json:"ticker"`
	Value     string `json:"value"`
	Change    string `json:"change"`
	Pct       string `json:"pct"`
	Dir       string `json:"dir"`
	Note      string `json:"note"`
	NoteColor string `json:"noteColor,omitempty"`
}

// QuickStat is a sidebar key metric.
type QuickStat struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	Change    string `json:"change"`
	ChangeDir string `json:"changeDir"`
}

// MarketDriver describes a macro catalyst card on the Overview page.
type MarketDriver struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Color string `json:"color"`
	Text  string `json:"text"`
}

// ChartData holds a labeled time series for line and bar charts.
type ChartData struct {
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
}

// CryptoChartData holds a labeled series with per-bar colors.
type CryptoChartData struct {
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
	Colors []string  `json:"colors"`
}

// OverviewResponse is the payload for GET /api/v1/overview.
type OverviewResponse struct {
	Header          HeaderInfo      `json:"header"`
	Indices         []Index         `json:"indices"`
	QuickStats      []QuickStat     `json:"quickStats"`
	MarketDrivers   []MarketDriver  `json:"marketDrivers"`
	SP500ChartData  ChartData       `json:"sp500ChartData"`
	CryptoChartData CryptoChartData `json:"cryptoChartData"`
	IntlChartData   ChartData       `json:"intlChartData"`
}

// USSummary holds top-line US market summary statistics.
type USSummary struct {
	MarchReturn       string `json:"marchReturn"`
	MarchNote         string `json:"marchNote"`
	YoYReturn         string `json:"yoYReturn"`
	YoYNote           string `json:"yoYNote"`
	ConsumerSentiment string `json:"consumerSentiment"`
	SentimentNote     string `json:"sentimentNote"`
}

// USMover represents a single notable US equity mover.
type USMover struct {
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
	Move   string `json:"move"`
	Dir    string `json:"dir"`
	Driver string `json:"driver"`
}

// IntlIndex represents a single international index row.
type IntlIndex struct {
	Ticker string `json:"ticker"`
	Region string `json:"region"`
	Change string `json:"change"`
	Dir    string `json:"dir"`
	Driver string `json:"driver"`
}

// SectorData holds sector performance and outlook scores for the heat map chart,
// plus ranked overweight/underweight sector names derived from those scores.
type SectorData struct {
	Labels      []string  `json:"labels"`
	Momentum    []float64 `json:"momentum"`
	Outlook     []float64 `json:"outlook"`
	Overweight  []string  `json:"overweight"`  // top sectors by combined score
	Underweight []string  `json:"underweight"` // bottom sectors by combined score
}

// MarketsResponse is the payload for GET /api/v1/markets.
type MarketsResponse struct {
	USSummary   USSummary   `json:"usSummary"`
	USMovers    []USMover   `json:"usMovers"`
	IntlIndices []IntlIndex `json:"intlIndices"`
	SectorData  SectorData  `json:"sectorData"`
}

// CryptoAsset represents a single crypto asset card.
type CryptoAsset struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Ticker string `json:"ticker"`
	Price  string `json:"price"`
	Change string `json:"change"`
	Dir    string `json:"dir"`
	Bg     string `json:"bg"`
	Color  string `json:"color"`
	Meta   string `json:"meta"`
}

// CryptoSmall is a compact crypto price card.
type CryptoSmall struct {
	Ticker string `json:"ticker"`
	Price  string `json:"price"`
	Change string `json:"change"`
	Dir    string `json:"dir"`
}

// CryptoResponse is the payload for GET /api/v1/crypto.
type CryptoResponse struct {
	CryptoAssets   []CryptoAsset `json:"cryptoAssets"`
	CryptoSmall    []CryptoSmall `json:"cryptoSmall"`
	BtcHistoryData ChartData     `json:"btcHistoryData"`
	MarketContext  string        `json:"marketContext"`
}

// RecStat is a single stat row on a recommendation card.
type RecStat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Dir   string `json:"dir"`
}

// Recommendation is a single stock or asset pick.
type Recommendation struct {
	Ticker string    `json:"ticker"`
	Name   string    `json:"name"`
	Badge  string    `json:"badge"`
	Type   string    `json:"type"`
	Stats  []RecStat `json:"stats"`
	Thesis string    `json:"thesis"`
}

// RecommendationsResponse is the payload for GET /api/v1/recommendations.
type RecommendationsResponse struct {
	AISemis    []Recommendation `json:"aiSemis"`
	Disruptors []Recommendation `json:"disruptors"`
	Defense    []Recommendation `json:"defense"`
	Crypto     []Recommendation `json:"crypto"`
}

// ChangeMetric holds a formatted dollar amount and percentage change with
// a direction indicator ("up", "down", "flat").
type ChangeMetric struct {
	Amount string `json:"amount"`
	Pct    string `json:"pct"`
	Dir    string `json:"dir"`
}

// PortfolioSummary holds top-line metrics for a portfolio account.
type PortfolioSummary struct {
	AccountName   string       `json:"accountName"`
	AccountType   string       `json:"accountType"`
	Provider      string       `json:"provider"`
	CurrentValue  string       `json:"currentValue"`
	TotalCost     string       `json:"totalCost"`
	TotalGain     string       `json:"totalGain"`
	TotalGainPct  string       `json:"totalGainPct"`
	TotalGainDir  string       `json:"totalGainDir"`
	WeeklyChange  ChangeMetric `json:"weeklyChange"`
	MonthlyChange ChangeMetric `json:"monthlyChange"`
	YTDChange     ChangeMetric `json:"ytdChange"`
}

// PortfolioHolding represents a single position within a portfolio.
type PortfolioHolding struct {
	Ticker     string `json:"ticker"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Allocation string `json:"allocation"`
	Gain       string `json:"gain"`
	GainDir    string `json:"gainDir"`
}

// Contribution records a single contribution or deposit event.
type Contribution struct {
	Date   string `json:"date"`
	Amount string `json:"amount"`
	Type   string `json:"type"` // "employee", "employer", "rollover", "deposit"
	Note   string `json:"note"`
}

// PortfolioResponse is the payload for a personal portfolio endpoint.
// Connected is false when the Plaid account has not yet been linked;
// in that case all other fields are zero values.
// PerformanceData is monthly account balance history.
// ContribData is monthly contribution totals for the bar chart.
type PortfolioResponse struct {
	Connected       bool               `json:"connected"`
	Summary         PortfolioSummary   `json:"summary"`
	PerformanceData ChartData          `json:"performanceData"`
	ContribData     ChartData          `json:"contribData"`
	Holdings        []PortfolioHolding `json:"holdings"`
	Contributions   []Contribution     `json:"contributions"`
}
