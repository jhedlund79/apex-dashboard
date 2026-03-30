import { useState, useCallback, useEffect, useRef } from 'react'
import { usePlaidLink } from 'react-plaid-link'
import { marketService } from '../services/market'

interface PlaidConnectProps {
  slot: 'principal' | 'morganstanley' | 'fidelity' | 'sofi'
  institutionName: string
  onConnected: () => void
}

export const PlaidConnect = ({ slot, institutionName, onConnected }: PlaidConnectProps) => {
  const [linkToken, setLinkToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const onSuccess = useCallback(
    async (publicToken: string) => {
      try {
        setLoading(true)
        await marketService.plaidExchange(slot, publicToken)
        onConnected()
      } catch {
        setError('Failed to save connection. Please try again.')
        setLoading(false)
      }
    },
    [slot, onConnected],
  )

  const opened = useRef(false)

  const { open, ready } = usePlaidLink({
    token: linkToken ?? '',
    onSuccess,
    onExit: () => {
      setLoading(false)
      setLinkToken(null)
      opened.current = false
    },
  })

  // Open the modal exactly once after the link token is ready.
  useEffect(() => {
    if (linkToken && ready && !opened.current) {
      opened.current = true
      open()
    }
  }, [linkToken, ready, open])

  const handleConnect = async () => {
    try {
      setLoading(true)
      setError(null)
      opened.current = false
      const { linkToken: token } = await marketService.plaidLinkToken(slot)
      setLinkToken(token)
    } catch {
      setError('Failed to initialize Plaid. Check your API keys.')
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <div className="max-w-md">
        <div className="text-[48px] mb-6">🔗</div>
        <h2 className="text-[22px] font-bold mb-2">Connect {institutionName}</h2>
        <p className="text-[13px] text-content-secondary leading-relaxed mb-8">
          Connect your {institutionName} account via Plaid to view your real portfolio balance,
          holdings, performance history, and contribution records.
        </p>

        {error && (
          <div className="mb-6 p-3 bg-danger/10 border border-danger/20 rounded-lg text-[12px] text-danger">
            {error}
          </div>
        )}

        <button
          onClick={handleConnect}
          disabled={loading}
          className="px-8 py-3 bg-accent text-bg font-semibold rounded-xl text-[14px] hover:bg-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
        >
          {loading ? 'Opening Plaid…' : 'Connect with Plaid'}
        </button>

        <p className="mt-6 text-[11px] text-content-muted leading-relaxed">
          Plaid uses bank-level encryption. Your credentials are never stored on this server —
          only a read-only access token is saved locally.
        </p>
      </div>
    </div>
  )
}
