import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useMarkets } from '../hooks/useMarketData'
import International from './International'

vi.mock('../hooks/useMarketData', () => ({ useMarkets: vi.fn() }))
const mockedUseMarkets = useMarkets as ReturnType<typeof vi.fn>

const marketsData = {
  usSummary: { marchReturn: '', marchNote: '', yoYReturn: '', yoYNote: '', consumerSentiment: '', sentimentNote: '' },
  usMovers: [],
  intlIndices: [
    { ticker: 'NKY', region: 'Japan', change: '-1.2%', dir: 'down', driver: 'Yen strength' },
    { ticker: 'DAX', region: 'Germany', change: '+0.5%', dir: 'up', driver: 'ECB policy' },
  ],
  sectorData: { labels: [], momentum: [], outlook: [], overweight: [], underweight: [] },
}

beforeEach(() => vi.clearAllMocks())

describe('International', () => {
  it('shows loading state', () => {
    mockedUseMarkets.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<International />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseMarkets.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<International />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders international content', () => {
    mockedUseMarkets.mockReturnValue({ data: marketsData, loading: false, error: null, retry: vi.fn() })
    render(<International />)
    expect(screen.getByText('International Markets')).toBeInTheDocument()
    expect(screen.getByText('NKY')).toBeInTheDocument()
    expect(screen.getByText('DAX')).toBeInTheDocument()
    expect(screen.getByText('Yen strength')).toBeInTheDocument()
  })
})
