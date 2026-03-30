import { describe, it, expect, vi, beforeEach } from 'vitest'
import { marketService } from './market'

const mockFetch = vi.fn()
global.fetch = mockFetch

function mockOk(data: unknown) {
  mockFetch.mockResolvedValueOnce({
    ok: true,
    json: async () => ({ data }),
  })
}

function mockError(status = 500) {
  mockFetch.mockResolvedValueOnce({ ok: false, status })
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('marketService', () => {
  it('overview fetches /api/v1/overview', async () => {
    const payload = { header: { fearGreed: 'Fear', marketStatus: 'Open' } }
    mockOk(payload)
    const result = await marketService.overview()
    expect(result).toEqual(payload)
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/overview')
  })

  it('markets fetches /api/v1/markets', async () => {
    mockOk({ usMovers: [] })
    const result = await marketService.markets()
    expect(result).toEqual({ usMovers: [] })
  })

  it('crypto fetches /api/v1/crypto', async () => {
    mockOk({ cryptoAssets: [] })
    await marketService.crypto()
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/crypto')
  })

  it('recommendations fetches /api/v1/recommendations', async () => {
    mockOk({ aiSemis: [] })
    await marketService.recommendations()
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/recommendations')
  })

  it('principal fetches /api/v1/portfolio/principal', async () => {
    mockOk({ connected: false })
    const result = await marketService.principal()
    expect(result).toEqual({ connected: false })
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/portfolio/principal')
  })

  it('morganStanley fetches /api/v1/portfolio/morgan-stanley', async () => {
    mockOk({ connected: true })
    await marketService.morganStanley()
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/portfolio/morgan-stanley')
  })

  it('fidelity fetches /api/v1/portfolio/fidelity', async () => {
    mockOk({ connected: true })
    await marketService.fidelity()
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/portfolio/fidelity')
  })

  it('soFi fetches /api/v1/portfolio/sofi', async () => {
    mockOk({ connected: false })
    const result = await marketService.soFi()
    expect(result).toEqual({ connected: false })
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/portfolio/sofi')
  })

  it('throws on non-ok response', async () => {
    mockError(404)
    await expect(marketService.overview()).rejects.toThrow('API error 404')
  })

  it('plaidStatus fetches /api/v1/plaid/status', async () => {
    mockOk({ enabled: true, principal: false, morganstanley: false })
    const result = await marketService.plaidStatus()
    expect(result.enabled).toBe(true)
  })

  it('plaidLinkToken POSTs to /api/v1/plaid/link-token', async () => {
    mockOk({ linkToken: 'link-sandbox-abc' })
    const result = await marketService.plaidLinkToken('principal')
    expect(result.linkToken).toBe('link-sandbox-abc')
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/plaid/link-token',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ slot: 'principal' }) }),
    )
  })

  it('plaidExchange POSTs to /api/v1/plaid/exchange', async () => {
    mockOk({ success: true })
    const result = await marketService.plaidExchange('principal', 'public-token')
    expect(result.success).toBe(true)
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/plaid/exchange',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ slot: 'principal', publicToken: 'public-token' }) }),
    )
  })

  it('apiPost throws on non-ok response', async () => {
    mockError(400)
    await expect(marketService.plaidLinkToken('principal')).rejects.toThrow('API error 400')
  })

  it('simulation fetches /api/v1/simulation', async () => {
    mockOk({ cashBalance: 100_000, totalValue: 100_000 })
    await marketService.simulation()
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/simulation')
  })

  it('simulationQuote fetches /api/v1/simulation/quote with ticker', async () => {
    mockOk({ ticker: 'NVDA', price: 135 })
    const result = await marketService.simulationQuote('NVDA')
    expect(result.price).toBe(135)
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/simulation/quote?ticker=NVDA')
  })

  it('simulationTrade POSTs to /api/v1/simulation/trade', async () => {
    mockOk({ success: true, price: 135, total: 1350 })
    const result = await marketService.simulationTrade('NVDA', 'buy', 10)
    expect(result.success).toBe(true)
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/simulation/trade',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ ticker: 'NVDA', action: 'buy', shares: 10 }) }),
    )
  })

  it('simulationReset POSTs to /api/v1/simulation/reset', async () => {
    mockOk({ success: true })
    const result = await marketService.simulationReset()
    expect(result.success).toBe(true)
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/simulation/reset',
      expect.objectContaining({ method: 'POST' }),
    )
  })
})
