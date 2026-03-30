import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useSimulation } from '../hooks/useMarketData'
import Simulation from './Simulation'

vi.mock('../hooks/useMarketData', () => ({ useSimulation: vi.fn() }))
vi.mock('../services/market', () => ({
  marketService: {
    simulationQuote: vi.fn(),
    simulationTrade: vi.fn(),
    simulationReset: vi.fn(),
  },
}))

const mockedUseSimulation = useSimulation as ReturnType<typeof vi.fn>

const emptySummary = {
  cashBalance: 100_000,
  startingCash: 100_000,
  portfolioValue: 0,
  totalValue: 100_000,
  totalGain: 0,
  totalGainPct: 0,
  positions: [],
  transactions: [],
  performanceData: { labels: [], data: [] },
}

const withPosition = {
  ...emptySummary,
  cashBalance: 88_500,
  portfolioValue: 11_500,
  totalValue: 100_000,
  positions: [
    {
      ticker: 'NVDA',
      shares: 10,
      avgCost: 115,
      currentPrice: 120,
      currentValue: 1200,
      gain: 50,
      gainPct: 4.35,
      gainDir: 'up' as const,
    },
  ],
  transactions: [
    {
      id: 'txn_1',
      date: '2026-03-01',
      ticker: 'NVDA',
      action: 'buy' as const,
      shares: 10,
      price: 115,
      total: 1150,
    },
  ],
}

beforeEach(() => vi.clearAllMocks())

describe('Simulation', () => {
  it('shows loading state', () => {
    mockedUseSimulation.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Simulation />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseSimulation.mockReturnValue({ data: null, loading: false, error: new Error('fail'), retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders summary bar with total value', () => {
    mockedUseSimulation.mockReturnValue({ data: emptySummary, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getAllByText('$100,000.00').length).toBeGreaterThan(0)
    expect(screen.getByText('Total Value')).toBeInTheDocument()
  })

  it('shows empty positions message', () => {
    mockedUseSimulation.mockReturnValue({ data: emptySummary, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByText(/No open positions/)).toBeInTheDocument()
  })

  it('renders open positions', () => {
    mockedUseSimulation.mockReturnValue({ data: withPosition, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getAllByText('NVDA').length).toBeGreaterThan(0)
    expect(screen.getByText('Open Positions')).toBeInTheDocument()
  })

  it('renders transaction history', () => {
    mockedUseSimulation.mockReturnValue({ data: withPosition, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByText('Transaction History')).toBeInTheDocument()
    expect(screen.getByText('2026-03-01')).toBeInTheDocument()
  })

  it('renders trade panel', () => {
    mockedUseSimulation.mockReturnValue({ data: emptySummary, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByPlaceholderText('Ticker (e.g. NVDA)')).toBeInTheDocument()
    expect(screen.getByText('Quote')).toBeInTheDocument()
  })

  it('renders reset button', () => {
    mockedUseSimulation.mockReturnValue({ data: emptySummary, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByText('Reset Portfolio')).toBeInTheDocument()
  })

  it('renders section header', () => {
    mockedUseSimulation.mockReturnValue({ data: emptySummary, loading: false, error: null, retry: vi.fn() })
    render(<Simulation />)
    expect(screen.getByText('◉ Simulator')).toBeInTheDocument()
  })
})
