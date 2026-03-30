import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useMarkets } from '../hooks/useMarketData'
import USMarkets from './USMarkets'

vi.mock('../hooks/useMarketData', () => ({ useMarkets: vi.fn() }))
const mockedUseMarkets = useMarkets as ReturnType<typeof vi.fn>

const marketsData = {
  usSummary: {
    marchReturn: '-5.8%',
    marchNote: 'Worst month since...',
    yoYReturn: '+8.2%',
    yoYNote: 'Still positive',
    consumerSentiment: '57.9',
    sentimentNote: '3-year low',
  },
  usMovers: [
    { ticker: 'NVDA', name: 'NVIDIA', move: '+8.4%', dir: 'up', driver: 'AI demand' },
  ],
  intlIndices: [],
  sectorData: { labels: [], momentum: [], outlook: [], overweight: [], underweight: [] },
}

beforeEach(() => vi.clearAllMocks())

describe('USMarkets', () => {
  it('shows loading state', () => {
    mockedUseMarkets.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<USMarkets />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseMarkets.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<USMarkets />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders US Markets content', () => {
    mockedUseMarkets.mockReturnValue({ data: marketsData, loading: false, error: null, retry: vi.fn() })
    render(<USMarkets />)
    expect(screen.getByText('US Markets')).toBeInTheDocument()
    expect(screen.getByText('-5.8%')).toBeInTheDocument()
    expect(screen.getByText('+8.2%')).toBeInTheDocument()
  })

  it('renders movers table', () => {
    mockedUseMarkets.mockReturnValue({ data: marketsData, loading: false, error: null, retry: vi.fn() })
    render(<USMarkets />)
    expect(screen.getByText('NVDA')).toBeInTheDocument()
    expect(screen.getByText('AI demand')).toBeInTheDocument()
  })
})
