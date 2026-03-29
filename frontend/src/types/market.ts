export interface HeaderInfo {
  fearGreed: string
  marketStatus: string
}

export interface Index {
  ticker: string
  value: string
  change: string
  pct: string
  dir: 'up' | 'down' | 'neutral'
  note: string
  noteColor?: string
}

export interface QuickStat {
  label: string
  value: string
  change: string
  changeDir: 'up' | 'down' | 'neutral'
}

export interface MarketDriver {
  icon: string
  title: string
  color: string
  text: string
}

export interface ChartData {
  labels: string[]
  data: number[]
}

export interface CryptoChartData {
  labels: string[]
  data: number[]
  colors: string[]
}

export interface OverviewResponse {
  header: HeaderInfo
  indices: Index[]
  quickStats: QuickStat[]
  marketDrivers: MarketDriver[]
  sp500ChartData: ChartData
  cryptoChartData: CryptoChartData
  intlChartData: ChartData
}

export interface USSummary {
  marchReturn: string
  marchNote: string
  yoYReturn: string
  yoYNote: string
  consumerSentiment: string
  sentimentNote: string
}

export interface USMover {
  ticker: string
  name: string
  move: string
  dir: 'up' | 'down' | 'neutral'
  driver: string
}

export interface IntlIndex {
  ticker: string
  region: string
  change: string
  dir: 'up' | 'down' | 'neutral'
  driver: string
}

export interface SectorData {
  labels: string[]
  momentum: number[]
  outlook: number[]
  overweight: string[]
  underweight: string[]
}

export interface MarketsResponse {
  usSummary: USSummary
  usMovers: USMover[]
  intlIndices: IntlIndex[]
  sectorData: SectorData
}

export interface CryptoAsset {
  symbol: string
  name: string
  ticker: string
  price: string
  change: string
  dir: 'up' | 'down' | 'neutral'
  bg: string
  color: string
  meta: string
}

export interface CryptoSmall {
  ticker: string
  price: string
  change: string
  dir: 'up' | 'down' | 'neutral'
}

export interface CryptoResponse {
  cryptoAssets: CryptoAsset[]
  cryptoSmall: CryptoSmall[]
  btcHistoryData: ChartData
  marketContext: string
}

export interface RecStat {
  label: string
  value: string
  dir: string
}

export interface Recommendation {
  ticker: string
  name: string
  badge: string
  type: 'strong-buy' | 'buy' | 'speculative'
  stats: RecStat[]
  thesis: string
}

export interface RecommendationsResponse {
  aiSemis: Recommendation[]
  disruptors: Recommendation[]
  defense: Recommendation[]
  crypto: Recommendation[]
}
