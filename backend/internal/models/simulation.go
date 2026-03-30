package models

// SimulationAccount is the persisted data format stored in simulation-data.json.
type SimulationAccount struct {
	CashBalance  float64              `json:"cashBalance"`
	StartingCash float64              `json:"startingCash"`
	Positions    []SimulationPosition `json:"positions"`
	Transactions []SimulationTx       `json:"transactions"`
	Snapshots    []PortfolioSnapshot  `json:"snapshots"`
}

// SimulationPosition is an open position in the simulation portfolio.
type SimulationPosition struct {
	Ticker       string  `json:"ticker"`
	Shares       float64 `json:"shares"`
	AvgCost      float64 `json:"avgCost"`
	PurchaseDate string  `json:"purchaseDate"`
}

// SimulationTx records a buy or sell transaction.
type SimulationTx struct {
	ID     string  `json:"id"`
	Date   string  `json:"date"`
	Ticker string  `json:"ticker"`
	Action string  `json:"action"` // "buy" | "sell"
	Shares float64 `json:"shares"`
	Price  float64 `json:"price"`
	Total  float64 `json:"total"`
}

// PortfolioSnapshot records total portfolio value at a point in time.
type PortfolioSnapshot struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// SimulationSummaryResponse is the response for GET /api/v1/simulation.
type SimulationSummaryResponse struct {
	CashBalance     float64                      `json:"cashBalance"`
	StartingCash    float64                      `json:"startingCash"`
	PortfolioValue  float64                      `json:"portfolioValue"`
	TotalValue      float64                      `json:"totalValue"`
	TotalGain       float64                      `json:"totalGain"`
	TotalGainPct    float64                      `json:"totalGainPct"`
	Positions       []SimulationPositionResponse `json:"positions"`
	Transactions    []SimulationTx               `json:"transactions"`
	PerformanceData ChartData                    `json:"performanceData"`
}

// SimulationPositionResponse enriches a position with live pricing.
type SimulationPositionResponse struct {
	Ticker       string  `json:"ticker"`
	Shares       float64 `json:"shares"`
	AvgCost      float64 `json:"avgCost"`
	CurrentPrice float64 `json:"currentPrice"`
	CurrentValue float64 `json:"currentValue"`
	Gain         float64 `json:"gain"`
	GainPct      float64 `json:"gainPct"`
	GainDir      string  `json:"gainDir"` // "up" | "down" | "flat"
}

// QuoteResponse is the response for GET /api/v1/simulation/quote.
type QuoteResponse struct {
	Ticker    string  `json:"ticker"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	Dir       string  `json:"dir"` // "up" | "down" | "neutral"
}

// TradeRequest is the body for POST /api/v1/simulation/trade.
type TradeRequest struct {
	Ticker string  `json:"ticker"`
	Action string  `json:"action"` // "buy" | "sell"
	Shares float64 `json:"shares"`
}

// TradeResponse is the response for POST /api/v1/simulation/trade.
type TradeResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	Price       float64 `json:"price"`
	Total       float64 `json:"total"`
	CashBalance float64 `json:"cashBalance"`
}
