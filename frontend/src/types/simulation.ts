export interface SimulationPosition {
  ticker: string
  shares: number
  avgCost: number
  currentPrice: number
  currentValue: number
  gain: number
  gainPct: number
  gainDir: 'up' | 'down' | 'flat'
}

export interface SimulationTx {
  id: string
  date: string
  ticker: string
  action: 'buy' | 'sell'
  shares: number
  price: number
  total: number
}

export interface SimulationSummary {
  cashBalance: number
  startingCash: number
  portfolioValue: number
  totalValue: number
  totalGain: number
  totalGainPct: number
  positions: SimulationPosition[]
  transactions: SimulationTx[]
  performanceData: { labels: string[]; data: number[] }
}

export interface QuoteResponse {
  ticker: string
  price: number
  change: number
  changePct: number
  dir: 'up' | 'down' | 'neutral'
}

export interface TradeResponse {
  success: boolean
  message: string
  price: number
  total: number
  cashBalance: number
}
