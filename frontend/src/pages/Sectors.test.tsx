import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useMarkets } from '../hooks/useMarketData'
import Sectors from './Sectors'

vi.mock('../hooks/useMarketData', () => ({ useMarkets: vi.fn() }))
const mockedUseMarkets = useMarkets as ReturnType<typeof vi.fn>

const marketsData = {
  usSummary: { marchReturn: '', marchNote: '', yoYReturn: '', yoYNote: '', consumerSentiment: '', sentimentNote: '' },
  usMovers: [],
  intlIndices: [],
  sectorData: {
    labels: ['Tech', 'Energy'],
    momentum: [8, 5],
    outlook: [9, 4],
    overweight: ['Tech', 'Healthcare'],
    underweight: ['Real Estate'],
  },
}

beforeEach(() => vi.clearAllMocks())

describe('Sectors', () => {
  it('shows loading state', () => {
    mockedUseMarkets.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Sectors />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseMarkets.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Sectors />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders sector content', () => {
    mockedUseMarkets.mockReturnValue({ data: marketsData, loading: false, error: null, retry: vi.fn() })
    render(<Sectors />)
    expect(screen.getByText('Sector Analysis')).toBeInTheDocument()
  })

  it('renders overweight and underweight lists', () => {
    mockedUseMarkets.mockReturnValue({ data: marketsData, loading: false, error: null, retry: vi.fn() })
    render(<Sectors />)
    expect(screen.getByText('Tech')).toBeInTheDocument()
    expect(screen.getByText('Healthcare')).toBeInTheDocument()
    expect(screen.getByText('Real Estate')).toBeInTheDocument()
  })
})
