import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { usePrincipal } from '../hooks/useMarketData'
import Principal from './Principal'

vi.mock('../hooks/useMarketData', () => ({ usePrincipal: vi.fn() }))
vi.mock('../services/market', () => ({
  marketService: { plaidLinkToken: vi.fn(), plaidExchange: vi.fn() },
}))
const mockedUsePrincipal = usePrincipal as ReturnType<typeof vi.fn>

const portfolioData = {
  connected: true,
  summary: {
    accountName: 'My 401k',
    accountType: '401(k)',
    provider: 'Principal Financial Group',
    currentValue: '$125,000',
    totalCost: '$100,000',
    totalGain: '+$25,000',
    totalGainPct: '+25.0%',
    totalGainDir: 'up' as const,
    weeklyChange: { amount: '+$500', pct: '+0.4%', dir: 'up' as const },
    monthlyChange: { amount: '-$2,000', pct: '-1.6%', dir: 'down' as const },
    ytdChange: { amount: '-$8,000', pct: '-6.0%', dir: 'down' as const },
  },
  performanceData: { labels: ['Jan', 'Feb'], data: [110000, 125000] },
  contribData: { labels: ['Jan', 'Feb'], data: [1200, 1200] },
  holdings: [
    { ticker: 'VIIIX', name: 'Vanguard Inst Index', value: '$75,000', allocation: '60.0%', gain: '+$15,000', gainDir: 'up' as const },
  ],
  contributions: [
    { date: 'Jan 15, 2025', amount: '$1,200', type: 'employee' as const, note: 'Employee Contribution' },
  ],
}

beforeEach(() => vi.clearAllMocks())

describe('Principal', () => {
  it('shows loading state', () => {
    mockedUsePrincipal.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Principal />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUsePrincipal.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Principal />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows PlaidConnect when not connected', () => {
    mockedUsePrincipal.mockReturnValue({
      data: { ...portfolioData, connected: false },
      loading: false,
      error: null,
      retry: vi.fn(),
    })
    render(<Principal />)
    expect(screen.getByRole('button', { name: /connect with plaid/i })).toBeInTheDocument()
  })

  it('renders portfolio content when connected', () => {
    mockedUsePrincipal.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Principal />)
    expect(screen.getByText('Principal Financial Group · 401(k)')).toBeInTheDocument()
    expect(screen.getByText('$125,000')).toBeInTheDocument()
    expect(screen.getByText('My 401k')).toBeInTheDocument()
  })

  it('renders holdings table', () => {
    mockedUsePrincipal.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Principal />)
    expect(screen.getByText('VIIIX')).toBeInTheDocument()
    expect(screen.getByText('60.0%')).toBeInTheDocument()
  })

  it('renders contribution history', () => {
    mockedUsePrincipal.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Principal />)
    expect(screen.getByText('Jan 15, 2025')).toBeInTheDocument()
    expect(screen.getByText('$1,200')).toBeInTheDocument()
  })

  it('renders change metrics', () => {
    mockedUsePrincipal.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Principal />)
    expect(screen.getByText('Weekly Change')).toBeInTheDocument()
    expect(screen.getByText('Monthly Change')).toBeInTheDocument()
    expect(screen.getByText('YTD Change')).toBeInTheDocument()
  })
})
