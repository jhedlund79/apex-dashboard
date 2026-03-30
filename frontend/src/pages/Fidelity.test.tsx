import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useFidelity } from '../hooks/useMarketData'
import Fidelity from './Fidelity'

vi.mock('../hooks/useMarketData', () => ({ useFidelity: vi.fn() }))
vi.mock('../services/market', () => ({
  marketService: { plaidLinkToken: vi.fn(), plaidExchange: vi.fn() },
}))
const mockedUseFidelity = useFidelity as ReturnType<typeof vi.fn>

const portfolioData = {
  connected: true,
  summary: {
    accountName: 'Individual Brokerage',
    accountType: 'Brokerage',
    provider: 'Fidelity',
    currentValue: '$43,218.74',
    totalCost: '$35,000.00',
    totalGain: '+$8,218.74',
    totalGainPct: '+23.48%',
    totalGainDir: 'up' as const,
    weeklyChange: { amount: '-$612.30', pct: '-1.40%', dir: 'down' as const },
    monthlyChange: { amount: '-$1,840.50', pct: '-4.08%', dir: 'down' as const },
    ytdChange: { amount: '-$3,210.00', pct: '-6.91%', dir: 'down' as const },
  },
  performanceData: { labels: ['Jan', 'Feb'], data: [40000, 43218] },
  contribData: { labels: ['Jan', 'Feb'], data: [0, 5000] },
  holdings: [
    { ticker: 'FSKAX', name: 'Fidelity Total Market Index', value: '$13,829.99', allocation: '32.0%', gain: '+$2,829.99', gainDir: 'up' as const },
  ],
  contributions: [
    { date: 'Jan 2, 2025', amount: '$5,000.00', type: 'deposit' as const, note: 'Annual contribution' },
  ],
}

beforeEach(() => vi.clearAllMocks())

describe('Fidelity', () => {
  it('shows loading state', () => {
    mockedUseFidelity.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Fidelity />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseFidelity.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Fidelity />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows PlaidConnect when not connected', () => {
    mockedUseFidelity.mockReturnValue({
      data: { ...portfolioData, connected: false },
      loading: false,
      error: null,
      retry: vi.fn(),
    })
    render(<Fidelity />)
    expect(screen.getByRole('button', { name: /connect with plaid/i })).toBeInTheDocument()
  })

  it('renders portfolio content when connected', () => {
    mockedUseFidelity.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Fidelity />)
    expect(screen.getByText('Fidelity · Brokerage')).toBeInTheDocument()
    expect(screen.getByText('$43,218.74')).toBeInTheDocument()
    expect(screen.getByText('Individual Brokerage')).toBeInTheDocument()
  })

  it('renders holdings table', () => {
    mockedUseFidelity.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Fidelity />)
    expect(screen.getByText('FSKAX')).toBeInTheDocument()
    expect(screen.getByText('32.0%')).toBeInTheDocument()
  })

  it('renders deposit history', () => {
    mockedUseFidelity.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Fidelity />)
    expect(screen.getByText('Jan 2, 2025')).toBeInTheDocument()
    expect(screen.getByText('$5,000.00')).toBeInTheDocument()
  })

  it('renders change metrics', () => {
    mockedUseFidelity.mockReturnValue({ data: portfolioData, loading: false, error: null, retry: vi.fn() })
    render(<Fidelity />)
    expect(screen.getByText('Weekly Change')).toBeInTheDocument()
    expect(screen.getByText('Monthly Change')).toBeInTheDocument()
    expect(screen.getByText('YTD Change')).toBeInTheDocument()
    expect(screen.getByText('Total Return')).toBeInTheDocument()
  })
})
