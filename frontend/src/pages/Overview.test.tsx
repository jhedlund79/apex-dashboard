import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useOverview } from '../hooks/useMarketData'
import Overview from './Overview'

vi.mock('../hooks/useMarketData', () => ({ useOverview: vi.fn() }))
const mockedUseOverview = useOverview as ReturnType<typeof vi.fn>

const overviewData = {
  header: { fearGreed: 'Fear', marketStatus: 'Open' },
  indices: [
    { ticker: 'SPY', value: '500', change: '+1', pct: '+0.2%', dir: 'up', note: 'note', noteColor: '' },
  ],
  quickStats: [],
  marketDrivers: [
    { icon: '📉', title: 'Tariffs', color: 'danger', text: 'Impact on markets' },
  ],
  sp500ChartData: { labels: ['Jan'], data: [4500] },
  cryptoChartData: { labels: ['BTC'], data: [5], colors: ['#f7931a'] },
  intlChartData: { labels: ['Nikkei'], data: [-1] },
}

beforeEach(() => vi.clearAllMocks())

describe('Overview', () => {
  it('shows loading state', () => {
    mockedUseOverview.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Overview />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseOverview.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Overview />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows error state when data is null', () => {
    mockedUseOverview.mockReturnValue({ data: null, loading: false, error: null, retry: vi.fn() })
    render(<Overview />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders overview content when data is loaded', () => {
    mockedUseOverview.mockReturnValue({ data: overviewData, loading: false, error: null, retry: vi.fn() })
    render(<Overview />)
    expect(screen.getByText('Market Overview')).toBeInTheDocument()
    expect(screen.getByText('SPY')).toBeInTheDocument()
    expect(screen.getByText(/Tariffs/)).toBeInTheDocument()
  })

  it('renders index values', () => {
    mockedUseOverview.mockReturnValue({ data: overviewData, loading: false, error: null, retry: vi.fn() })
    render(<Overview />)
    expect(screen.getByText('500')).toBeInTheDocument()
  })
})
