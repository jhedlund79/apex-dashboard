import { useState, useEffect, useCallback } from 'react'
import { marketService } from '../services/market'
import type {
  OverviewResponse,
  MarketsResponse,
  CryptoResponse,
  RecommendationsResponse,
} from '../types/market'

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
