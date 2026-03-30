import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useSoFi } from '../hooks/useMarketData'
import SoFi from './SoFi'

vi.mock('../hooks/useMarketData', () => ({ useSoFi: vi.fn() }))
vi.mock('../services/market', () => ({
  marketService: { plaidLinkToken: vi.fn(), plaidExchange: vi.fn() },
}))
const mockedUseSoFi = useSoFi as ReturnType<typeof vi.fn>

const portfolioData = {
  connected: true,
  summary: {
    accountName: 'Active Invest',
    accountType: 'Brokerage',
    provider: 'SoFi',
    currentValue: '$8,142.60',
    totalCost: '$7,500.00',
    totalGain: '+$642.60',
    totalGainPct: '+8.57%',
    totalGainDir: 'up' as const,
    weeklyChange: { amount: '-$184.20', pct: '-2.21%', dir: 'down' as const },
    monthlyChange: { amount: '-$390.80', pct: '-4.58%', dir: 'down' as const },
    ytdChange: { amount: '-$607.40', pct: '-6.94%', dir: 'down' as const },
  },
  performanceData: { labels: ['Jan', 'Feb'], data: [7800, 8142] },
  contribData: { labels: ['Jan', 'Feb'], data: [250, 250] },
  holdings: [
    { ticker: 'VTI', name: 'Vanguard Total Stock Market ETF', value: '$2,442.78', allocation: '30.0%', gain: '+$192.78', gainDir: 'up' as const },
  ],
  contributions: [
    { date: 'Mar 1, 2025', amount: '$250.00', type: 'deposit' as const, note: 'Monthly auto-invest' },
  ],
}

beforeEach(() => vi.clearAllMocks())

describe('SoFi', () => {
  it('shows loading state', () => {
    mockedUseSoFi.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<SoFi />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseSoFi.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<SoFi />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows PlaidConnect when not connected', () => {
    mockedUseSoFi.mockReturnValue({
      data: { ...portfolioData, connected: false },
      loading: false,
      error: null,
      retry: vi.fn(),
    })
    render(<SoFi />)
    expect(screen.getByRole('button', { name: /connect with plaid/i })).toBeInTheDocument()
  })

  it('renders portfolio content when connected', () => {
    mockedUseSoFi.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<SoFi />)
    expect(screen.getByText('SoFi · Brokerage')).toBeInTheDocument()
    expect(screen.getByText('$8,142.60')).toBeInTheDocument()
    expect(screen.getByText('Active Invest')).toBeInTheDocument()
  })

  it('renders holdings table', () => {
    mockedUseSoFi.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<SoFi />)
    expect(screen.getByText('VTI')).toBeInTheDocument()
    expect(screen.getByText('30.0%')).toBeInTheDocument()
  })

  it('renders deposit history', () => {
    mockedUseSoFi.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<SoFi />)
    expect(screen.getByText('Mar 1, 2025')).toBeInTheDocument()
    expect(screen.getByText('$250.00')).toBeInTheDocument()
  })

  it('renders change metrics', () => {
    mockedUseSoFi.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<SoFi />)
    expect(screen.getByText('Weekly Change')).toBeInTheDocument()
    expect(screen.getByText('Monthly Change')).toBeInTheDocument()
    expect(screen.getByText('YTD Change')).toBeInTheDocument()
    expect(screen.getByText('Total Return')).toBeInTheDocument()
  })
})
