import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import {
  useOverview,
  useMarkets,
  useCrypto,
  useRecommendations,
  usePrincipal,
  useMorganStanley,
  useFidelity,
  useSoFi,
  useSimulation,
} from './useMarketData'
import { marketService } from '../services/market'

vi.mock('../services/market', () => ({
  marketService: {
    overview: vi.fn(),
    markets: vi.fn(),
    crypto: vi.fn(),
    recommendations: vi.fn(),
    principal: vi.fn(),
    morganStanley: vi.fn(),
    fidelity: vi.fn(),
    soFi: vi.fn(),
    simulation: vi.fn(),
  },
}))

const mocked = marketService as unknown as {
  overview: ReturnType<typeof vi.fn>
  markets: ReturnType<typeof vi.fn>
  crypto: ReturnType<typeof vi.fn>
  recommendations: ReturnType<typeof vi.fn>
  principal: ReturnType<typeof vi.fn>
  morganStanley: ReturnType<typeof vi.fn>
  fidelity: ReturnType<typeof vi.fn>
  soFi: ReturnType<typeof vi.fn>
  simulation: ReturnType<typeof vi.fn>
}

beforeEach(() => vi.clearAllMocks())

describe('useOverview', () => {
  it('returns data on success', async () => {
    const data = { header: { fearGreed: 'Greed', marketStatus: 'Open' } }
    mocked.overview.mockResolvedValueOnce(data)
    const { result } = renderHook(() => useOverview())
    expect(result.current.loading).toBe(true)
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data).toEqual(data)
    expect(result.current.error).toBeNull()
  })

  it('sets error on failure', async () => {
    mocked.overview.mockRejectedValueOnce(new Error('network fail'))
    const { result } = renderHook(() => useOverview())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.error?.message).toBe('network fail')
    expect(result.current.data).toBeNull()
  })

  it('wraps non-Error rejections', async () => {
    mocked.overview.mockRejectedValueOnce('string error')
    const { result } = renderHook(() => useOverview())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.error).toBeInstanceOf(Error)
  })

  it('retry re-fetches data', async () => {
    mocked.overview.mockResolvedValue({ header: {} })
    const { result } = renderHook(() => useOverview())
    await waitFor(() => expect(result.current.loading).toBe(false))
    result.current.retry()
    await waitFor(() => expect(mocked.overview).toHaveBeenCalledTimes(2))
  })
})

describe('useMarkets', () => {
  it('fetches markets data', async () => {
    mocked.markets.mockResolvedValueOnce({ usMovers: [] })
    const { result } = renderHook(() => useMarkets())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data).toEqual({ usMovers: [] })
  })
})

describe('useCrypto', () => {
  it('fetches crypto data', async () => {
    mocked.crypto.mockResolvedValueOnce({ cryptoAssets: [] })
    const { result } = renderHook(() => useCrypto())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data).toEqual({ cryptoAssets: [] })
  })
})

describe('useRecommendations', () => {
  it('fetches recommendations data', async () => {
    mocked.recommendations.mockResolvedValueOnce({ aiSemis: [] })
    const { result } = renderHook(() => useRecommendations())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data).toEqual({ aiSemis: [] })
  })
})

describe('usePrincipal', () => {
  it('fetches principal portfolio', async () => {
    mocked.principal.mockResolvedValueOnce({ connected: false })
    const { result } = renderHook(() => usePrincipal())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data?.connected).toBe(false)
  })
})

describe('useMorganStanley', () => {
  it('fetches Morgan Stanley portfolio', async () => {
    mocked.morganStanley.mockResolvedValueOnce({ connected: true })
    const { result } = renderHook(() => useMorganStanley())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data?.connected).toBe(true)
  })
})

describe('useFidelity', () => {
  it('fetches Fidelity portfolio', async () => {
    mocked.fidelity.mockResolvedValueOnce({ connected: true })
    const { result } = renderHook(() => useFidelity())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data?.connected).toBe(true)
  })
})

describe('useSoFi', () => {
  it('fetches SoFi portfolio', async () => {
    mocked.soFi.mockResolvedValueOnce({ connected: false })
    const { result } = renderHook(() => useSoFi())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data?.connected).toBe(false)
  })
})

describe('useSimulation', () => {
  it('fetches simulation summary', async () => {
    const data = { cashBalance: 100_000, startingCash: 100_000, totalValue: 100_000, positions: [] }
    mocked.simulation.mockResolvedValueOnce(data)
    const { result } = renderHook(() => useSimulation())
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.data?.cashBalance).toBe(100_000)
  })
})
