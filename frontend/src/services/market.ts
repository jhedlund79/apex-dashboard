import type {
  OverviewResponse,
  MarketsResponse,
  CryptoResponse,
  RecommendationsResponse,
  PortfolioResponse,
  PlaidStatus,
} from '../types/market'

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`/api/v1${path}`)
  if (!res.ok) throw new Error(`API error ${res.status}: ${path}`)
  const { data } = (await res.json()) as { data: T }
  return data
}

async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`API error ${res.status}: ${path}`)
  const { data } = (await res.json()) as { data: T }
  return data
}

export const marketService = {
  overview: (): Promise<OverviewResponse> => apiFetch('/overview'),
  markets: (): Promise<MarketsResponse> => apiFetch('/markets'),
  crypto: (): Promise<CryptoResponse> => apiFetch('/crypto'),
  recommendations: (): Promise<RecommendationsResponse> => apiFetch('/recommendations'),
  principal: (): Promise<PortfolioResponse> => apiFetch('/portfolio/principal'),
  morganStanley: (): Promise<PortfolioResponse> => apiFetch('/portfolio/morgan-stanley'),
  fidelity: (): Promise<PortfolioResponse> => apiFetch('/portfolio/fidelity'),
  soFi: (): Promise<PortfolioResponse> => apiFetch('/portfolio/sofi'),

  plaidStatus: (): Promise<PlaidStatus> => apiFetch('/plaid/status'),
  plaidLinkToken: (slot: string): Promise<{ linkToken: string }> =>
    apiPost('/plaid/link-token', { slot }),
  plaidExchange: (slot: string, publicToken: string): Promise<{ success: boolean }> =>
    apiPost('/plaid/exchange', { slot, publicToken }),
}
