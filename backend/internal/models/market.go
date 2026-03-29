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
