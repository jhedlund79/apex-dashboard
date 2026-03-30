import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useRecommendations } from '../hooks/useMarketData'
import Recommendations from './Recommendations'

vi.mock('../hooks/useMarketData', () => ({ useRecommendations: vi.fn() }))
const mockedUseRecommendations = useRecommendations as ReturnType<typeof vi.fn>

const rec = {
  ticker: 'NVDA',
  name: 'NVIDIA Corporation',
  badge: 'Strong Buy',
  type: 'strong-buy' as const,
  stats: [{ label: 'Target', value: '$180', dir: 'up' }],
  thesis: 'AI demand remains strong',
}

const recsData = {
  aiSemis: [rec],
  disruptors: [],
  defense: [],
  crypto: [],
}

beforeEach(() => vi.clearAllMocks())

describe('Recommendations', () => {
  it('shows loading state', () => {
    mockedUseRecommendations.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Recommendations />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseRecommendations.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Recommendations />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders recommendations', () => {
    mockedUseRecommendations.mockReturnValue({ data: recsData, loading: false, error: null, retry: vi.fn() })
    render(<Recommendations />)
    expect(screen.getByText('NVDA')).toBeInTheDocument()
    expect(screen.getByText('NVIDIA Corporation')).toBeInTheDocument()
    expect(screen.getByText('Strong Buy')).toBeInTheDocument()
  })

  it('renders thesis text', () => {
    mockedUseRecommendations.mockReturnValue({ data: recsData, loading: false, error: null, retry: vi.fn() })
    render(<Recommendations />)
    expect(screen.getByText('AI demand remains strong')).toBeInTheDocument()
  })

  it('renders section headers', () => {
    mockedUseRecommendations.mockReturnValue({ data: recsData, loading: false, error: null, retry: vi.fn() })
    render(<Recommendations />)
    expect(screen.getByText(/Aggressive Growth Picks/)).toBeInTheDocument()
    expect(screen.getByText('AI & Semiconductor Growth')).toBeInTheDocument()
  })
})
