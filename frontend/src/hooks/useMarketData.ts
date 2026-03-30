import { useState, useEffect, useCallback } from 'react'
import { marketService } from '../services/market'
import type {
  OverviewResponse,
  MarketsResponse,
  CryptoResponse,
  RecommendationsResponse,
  PortfolioResponse,
} from '../types/market'
import type { SimulationSummary } from '../types/simulation'

interface ApiState<T> {
  data: T | null
  loading: boolean
  error: Error | null
  retry: () => void
}

function useApiCall<T>(fetcher: () => Promise<T>): ApiState<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    fetcher()
      .then(setData)
      .catch((err: unknown) => setError(err instanceof Error ? err : new Error(String(err))))
      .finally(() => setLoading(false))
  }, [fetcher])

  useEffect(load, [load])

  return { data, loading, error, retry: load }
}

export const useOverview = (): ApiState<OverviewResponse> =>
  useApiCall(marketService.overview)

export const useMarkets = (): ApiState<MarketsResponse> =>
  useApiCall(marketService.markets)

export const useCrypto = (): ApiState<CryptoResponse> =>
  useApiCall(marketService.crypto)

export const useRecommendations = (): ApiState<RecommendationsResponse> =>
  useApiCall(marketService.recommendations)

export const usePrincipal = (): ApiState<PortfolioResponse> =>
  useApiCall(marketService.principal)

export const useMorganStanley = (): ApiState<PortfolioResponse> =>
  useApiCall(marketService.morganStanley)

export const useFidelity = (): ApiState<PortfolioResponse> =>
  useApiCall(marketService.fidelity)

export const useSoFi = (): ApiState<PortfolioResponse> =>
  useApiCall(marketService.soFi)

export const useSimulation = (): ApiState<SimulationSummary> =>
  useApiCall(marketService.simulation)
