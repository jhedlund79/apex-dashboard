import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { PlaidConnect } from './PlaidConnect'
import { marketService } from '../services/market'

vi.mock('../services/market', () => ({
  marketService: {
    plaidLinkToken: vi.fn(),
    plaidExchange: vi.fn(),
  },
}))

const mocked = marketService as unknown as {
  plaidLinkToken: ReturnType<typeof vi.fn>
  plaidExchange: ReturnType<typeof vi.fn>
}

beforeEach(() => vi.clearAllMocks())

describe('PlaidConnect', () => {
  it('renders connect button', () => {
    render(
      <PlaidConnect slot="principal" institutionName="Principal Financial Group" onConnected={vi.fn()} />,
    )
    expect(screen.getByRole('button', { name: /connect with plaid/i })).toBeInTheDocument()
  })

  it('renders institution name in heading', () => {
    render(
      <PlaidConnect slot="principal" institutionName="Principal Financial Group" onConnected={vi.fn()} />,
    )
    expect(screen.getByText(/Connect Principal Financial Group/)).toBeInTheDocument()
  })

  it('shows loading state while fetching link token', async () => {
    mocked.plaidLinkToken.mockImplementation(() => new Promise(() => {})) // never resolves
    render(
      <PlaidConnect slot="principal" institutionName="Principal Financial Group" onConnected={vi.fn()} />,
    )
    fireEvent.click(screen.getByRole('button', { name: /connect with plaid/i }))
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /opening plaid/i })).toBeDisabled(),
    )
  })

  it('shows error when plaidLinkToken fails', async () => {
    mocked.plaidLinkToken.mockRejectedValueOnce(new Error('bad keys'))
    render(
      <PlaidConnect slot="principal" institutionName="Principal Financial Group" onConnected={vi.fn()} />,
    )
    fireEvent.click(screen.getByRole('button', { name: /connect with plaid/i }))
    await waitFor(() =>
      expect(screen.getByText(/Failed to initialize Plaid/)).toBeInTheDocument(),
    )
  })

  it('passes slot to plaidLinkToken', async () => {
    mocked.plaidLinkToken.mockResolvedValueOnce({ linkToken: 'link-token' })
    render(
      <PlaidConnect slot="morganstanley" institutionName="Morgan Stanley" onConnected={vi.fn()} />,
    )
    fireEvent.click(screen.getByRole('button'))
    await waitFor(() => expect(mocked.plaidLinkToken).toHaveBeenCalledWith('morganstanley'))
  })

  it('renders privacy disclaimer text', () => {
    render(
      <PlaidConnect slot="principal" institutionName="Principal" onConnected={vi.fn()} />,
    )
    expect(screen.getByText(/bank-level encryption/i)).toBeInTheDocument()
  })
})
