import type {
  OverviewResponse,
  MarketsResponse,
  CryptoResponse,
  RecommendationsResponse,
} from '../types/market'

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`/api/v1${path}`)
  if (!res.ok) throw new Error(`API error ${res.status}: ${path}`)
  const { data } = (await res.json()) as { data: T }
  return data
}

export const marketService = {
  overview: (): Promise<OverviewResponse> => apiFetch('/overview'),
  markets: (): Promise<MarketsResponse> => apiFetch('/markets'),
  crypto: (): Promise<CryptoResponse> => apiFetch('/crypto'),
  recommendations: (): Promise<RecommendationsResponse> => apiFetch('/recommendations'),
}
