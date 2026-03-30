import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useCrypto } from '../hooks/useMarketData'
import Crypto from './Crypto'

vi.mock('../hooks/useMarketData', () => ({ useCrypto: vi.fn() }))
const mockedUseCrypto = useCrypto as ReturnType<typeof vi.fn>

const cryptoData = {
  cryptoAssets: [
    { symbol: '₿', name: 'Bitcoin', ticker: 'BTC', price: '$82,500', change: '-2.1%', dir: 'down', bg: '#f7931a20', color: '#f7931a', meta: 'Dominance: 62%' },
  ],
  cryptoSmall: [
    { ticker: 'LINK', price: '$12.40', change: '+1.5%', dir: 'up' },
  ],
  btcHistoryData: { labels: ['Jan', 'Feb'], data: [45000, 82000] },
  marketContext: 'Market cap: $2.4T',
}

beforeEach(() => vi.clearAllMocks())

describe('Crypto', () => {
  it('shows loading state', () => {
    mockedUseCrypto.mockReturnValue({ data: null, loading: true, error: null, retry: vi.fn() })
    const { container } = render(<Crypto />)
    expect(container.querySelector('.animate-shimmer')).toBeInTheDocument()
  })

  it('shows error state', () => {
    mockedUseCrypto.mockReturnValue({
      data: null, loading: false, error: new Error('fail'), retry: vi.fn(),
    })
    render(<Crypto />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders crypto content', () => {
    mockedUseCrypto.mockReturnValue({ data: cryptoData, loading: false, error: null, retry: vi.fn() })
    render(<Crypto />)
    expect(screen.getByText('Crypto Markets')).toBeInTheDocument()
    expect(screen.getByText('Bitcoin')).toBeInTheDocument()
    expect(screen.getByText('$82,500')).toBeInTheDocument()
  })

  it('renders small cap section', () => {
    mockedUseCrypto.mockReturnValue({ data: cryptoData, loading: false, error: null, retry: vi.fn() })
    render(<Crypto />)
    expect(screen.getByText('LINK')).toBeInTheDocument()
  })

  it('renders market context', () => {
    mockedUseCrypto.mockReturnValue({ data: cryptoData, loading: false, error: null, retry: vi.fn() })
    render(<Crypto />)
    expect(screen.getByText('Market cap: $2.4T')).toBeInTheDocument()
  })
})
