import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useMorganStanley } from '../hooks/useMarketData'
import MorganStanley from './MorganStanley'

vi.mock('../hooks/useMarketData', () => ({ useMorganStanley: vi.fn() }))
vi.mock('../services/market', () => ({
  marketService: { plaidLinkToken: vi.fn(), plaidExchange: vi.fn() },
}))
const mockedUseMorganStanley = useMorganStanley as ReturnType<typeof vi.fn>

const portfolioData = {
  connected: true,
  summary: {
    accountName: 'Brokerage Account',
    accountType: 'Investment',
    provider: 'Morgan Stanley',
    currentValue: '$250,000',
    totalCost: '$200,000',
    totalGain: '+$50,000',
    totalGainPct: '+25.0%',
    totalGainDir: 'up' as const,
    weeklyChange: { amount: '+$1,000', pct: '+0.4%', dir: 'up' as const },
    monthlyChange: { amount: '-$3,000', pct: '-1.2%', dir: 'down' as const },
    ytdChange: { amount: '-$10,000', pct: '-4.0%', dir: 'down' as const },
  },
  performanceData: { labels: ['Jan', 'Feb'], data: [230000, 250000] },
  contribData: { labels: ['Jan', 'Feb'], data: [5000, 0] },
  holdings: [
    { ticker: 'AAPL', name: 'Apple Inc', value: '$50,000', allocation: '20.0%', gain: '+$10,000', gainDir: 'up' as const },
  ],
  contributions: [
    { date: 'Jan 2, 2025', amount: '$5,000', type: 'deposit' as const, note: 'Wire Transfer' },
  ],
}

beforeEach(() => vi.clearAllMocks())

describe('MorganStanley', () => {
  it('shows loading state', () => {
    mockedUseMorganStanley.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<MorganStanley />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseMorganStanley.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<MorganStanley />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows PlaidConnect when not connected', () => {
    mockedUseMorganStanley.mockReturnValue({
      data: { ...portfolioData, connected: false },
      loading: false,
      error: null,
      retry: vi.fn(),
    })
    render(<MorganStanley />)
    expect(screen.getByRole('button', { name: /connect with plaid/i })).toBeInTheDocument()
  })

  it('renders portfolio content when connected', () => {
    mockedUseMorganStanley.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<MorganStanley />)
    expect(screen.getByText('Morgan Stanley · Investment')).toBeInTheDocument()
    expect(screen.getByText('$250,000')).toBeInTheDocument()
    expect(screen.getByText('Brokerage Account')).toBeInTheDocument()
  })

  it('renders holdings', () => {
    mockedUseMorganStanley.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<MorganStanley />)
    expect(screen.getByText('AAPL')).toBeInTheDocument()
  })

  it('renders change metrics', () => {
    mockedUseMorganStanley.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<MorganStanley />)
    expect(screen.getByText('Weekly Change')).toBeInTheDocument()
    expect(screen.getByText('Total Return')).toBeInTheDocument()
  })
})
