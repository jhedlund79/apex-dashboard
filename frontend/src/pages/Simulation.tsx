import { Line } from 'react-chartjs-2'
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, Tooltip, Filler,
} from 'chart.js'
import { useState, useCallback } from 'react'
import { useSimulation } from '../hooks/useMarketData'
import { marketService } from '../services/market'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Table, Th, Td, TickerCell } from './shared'
import type { SimulationPosition, SimulationTx, QuoteResponse } from '../types/simulation'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const axisStyle = {
  grid: { color: 'rgba(255,255,255,0.03)' },
  ticks: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } },
}

const fmt = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n)

const fmtPct = (n: number) => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`

const dirClass = (dir: string) =>
  dir === 'up' ? 'text-accent' : dir === 'down' ? 'text-danger' : 'text-content-primary'

// ── Holdings Table ────────────────────────────────────────────────────────────

const HoldingsTable = ({
  positions,
  onSell,
}: {
  positions: SimulationPosition[]
  onSell: (ticker: string, maxShares: number) => void
}) => (
  <div className="bg-surface border border-white/[0.04] rounded-card overflow-hidden mb-6 animate-fadeIn">
    <div className="p-5 pb-0">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-3.5">
        Open Positions
      </p>
    </div>
    {positions.length === 0 ? (
      <p className="px-5 pb-5 text-content-muted text-sm">No open positions. Use the trade form below to buy your first stock.</p>
    ) : (
      <Table>
        <thead>
          <tr>
            <Th>Ticker</Th>
            <Th>Shares</Th>
            <Th>Avg Cost</Th>
            <Th>Current Price</Th>
            <Th>Value</Th>
            <Th>Gain / Loss</Th>
            <Th>{''}</Th>
          </tr>
        </thead>
        <tbody>
          {positions.map(p => (
            <tr key={p.ticker} className="hover:bg-white/[0.02] transition-colors">
              <TickerCell>{p.ticker}</TickerCell>
              <Td className="font-mono">{p.shares.toFixed(4)}</Td>
              <Td className="font-mono">{fmt(p.avgCost)}</Td>
              <Td className="font-mono">{fmt(p.currentPrice)}</Td>
              <Td className="font-mono font-semibold">{fmt(p.currentValue)}</Td>
              <Td className={`font-mono ${dirClass(p.gainDir)}`}>
                {fmt(p.gain)} ({fmtPct(p.gainPct)})
              </Td>
              <Td>
                <button
                  onClick={() => onSell(p.ticker, p.shares)}
                  className="text-[11px] px-2.5 py-1 rounded border border-danger/40 text-danger hover:bg-danger/10 font-mono transition-colors"
                >
                  Sell
                </button>
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>
    )}
  </div>
)

// ── Transaction History ───────────────────────────────────────────────────────

const TxTable = ({ transactions }: { transactions: SimulationTx[] }) => (
  <div className="bg-surface border border-white/[0.04] rounded-card overflow-hidden mb-6 animate-fadeIn">
    <div className="p-5 pb-0">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-3.5">
        Transaction History
      </p>
    </div>
    {transactions.length === 0 ? (
      <p className="px-5 pb-5 text-content-muted text-sm">No transactions yet.</p>
    ) : (
      <Table>
        <thead>
          <tr>
            <Th>Date</Th>
            <Th>Action</Th>
            <Th>Ticker</Th>
            <Th>Shares</Th>
            <Th>Price</Th>
            <Th>Total</Th>
          </tr>
        </thead>
        <tbody>
          {transactions.slice(0, 20).map(tx => (
            <tr key={tx.id} className="hover:bg-white/[0.02] transition-colors">
              <Td className="font-mono text-content-muted">{tx.date}</Td>
              <Td>
                <span className={`text-[11px] px-2 py-0.5 rounded-full font-mono uppercase tracking-wide border ${
                  tx.action === 'buy'
                    ? 'bg-accent/10 text-accent border-accent/20'
                    : 'bg-danger/10 text-danger border-danger/20'
                }`}>
                  {tx.action}
                </span>
              </Td>
              <TickerCell>{tx.ticker}</TickerCell>
              <Td className="font-mono">{tx.shares.toFixed(4)}</Td>
              <Td className="font-mono">{fmt(tx.price)}</Td>
              <Td className="font-mono font-semibold">{fmt(tx.total)}</Td>
            </tr>
          ))}
        </tbody>
      </Table>
    )}
  </div>
)

// ── Trade Panel ───────────────────────────────────────────────────────────────

interface TradePanelProps {
  cashBalance: number
  prefillTicker?: string
  prefillAction?: 'buy' | 'sell'
  prefillMaxShares?: number
  onTraded: () => void
}

const TradePanel = ({
  cashBalance,
  prefillTicker = '',
  prefillAction = 'buy',
  prefillMaxShares,
  onTraded,
}: TradePanelProps) => {
  const [ticker, setTicker] = useState(prefillTicker)
  const [action, setAction] = useState<'buy' | 'sell'>(prefillAction)
  const [shares, setShares] = useState(prefillMaxShares ? String(prefillMaxShares) : '')
  const [quote, setQuote] = useState<QuoteResponse | null>(null)
  const [quoteLoading, setQuoteLoading] = useState(false)
  const [tradeLoading, setTradeLoading] = useState(false)
  const [message, setMessage] = useState<{ text: string; ok: boolean } | null>(null)

  const lookupQuote = useCallback(async () => {
    if (!ticker.trim()) return
    setQuoteLoading(true)
    setQuote(null)
    setMessage(null)
    try {
      const q = await marketService.simulationQuote(ticker.trim().toUpperCase())
      setQuote(q)
    } catch {
      setMessage({ text: 'Ticker not found or live data unavailable', ok: false })
    } finally {
      setQuoteLoading(false)
    }
  }, [ticker])

  const executeTrade = useCallback(async () => {
    const sharesNum = parseFloat(shares)
    if (!ticker.trim() || !sharesNum || sharesNum <= 0) {
      setMessage({ text: 'Enter a ticker and number of shares', ok: false })
      return
    }
    setTradeLoading(true)
    setMessage(null)
    try {
      const res = await marketService.simulationTrade(ticker.trim().toUpperCase(), action, sharesNum)
      setMessage({ text: `${action === 'buy' ? 'Bought' : 'Sold'} ${sharesNum} ${ticker.toUpperCase()} @ ${fmt(res.price)} · Cash: ${fmt(res.cashBalance)}`, ok: true })
      setShares('')
      setQuote(null)
      onTraded()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Trade failed'
      setMessage({ text: msg, ok: false })
    } finally {
      setTradeLoading(false)
    }
  }, [ticker, action, shares, onTraded])

  const estimatedCost = quote && parseFloat(shares) > 0
    ? parseFloat(shares) * quote.price
    : null

  return (
    <div className="bg-surface border border-white/[0.04] rounded-card p-5 mb-6 animate-fadeIn">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-4">Trade</p>

      <div className="flex flex-wrap gap-3 mb-4">
        {/* Ticker input + quote */}
        <div className="flex gap-2 flex-1 min-w-[220px]">
          <input
            type="text"
            placeholder="Ticker (e.g. NVDA)"
            value={ticker}
            onChange={e => { setTicker(e.target.value.toUpperCase()); setQuote(null) }}
            onKeyDown={e => e.key === 'Enter' && lookupQuote()}
            className="flex-1 bg-bg border border-white/[0.08] rounded-lg px-3 py-2 text-sm font-mono text-content-primary placeholder-content-muted focus:outline-none focus:border-accent/40 uppercase"
          />
          <button
            onClick={lookupQuote}
            disabled={quoteLoading || !ticker.trim()}
            className="px-4 py-2 text-sm font-mono bg-white/[0.04] border border-white/[0.08] rounded-lg text-content-secondary hover:text-accent hover:border-accent/30 disabled:opacity-40 transition-colors"
          >
            {quoteLoading ? '…' : 'Quote'}
          </button>
        </div>

        {/* Buy / Sell toggle */}
        <div className="flex rounded-lg overflow-hidden border border-white/[0.08]">
          <button
            onClick={() => setAction('buy')}
            className={`px-5 py-2 text-sm font-mono transition-colors ${action === 'buy' ? 'bg-accent/20 text-accent' : 'text-content-muted hover:text-content-secondary'}`}
          >
            Buy
          </button>
          <button
            onClick={() => setAction('sell')}
            className={`px-5 py-2 text-sm font-mono border-l border-white/[0.08] transition-colors ${action === 'sell' ? 'bg-danger/20 text-danger' : 'text-content-muted hover:text-content-secondary'}`}
          >
            Sell
          </button>
        </div>

        {/* Shares */}
        <input
          type="number"
          placeholder="Shares"
          value={shares}
          onChange={e => setShares(e.target.value)}
          min="0.0001"
          step="1"
          className="w-32 bg-bg border border-white/[0.08] rounded-lg px-3 py-2 text-sm font-mono text-content-primary placeholder-content-muted focus:outline-none focus:border-accent/40"
        />

        <button
          onClick={executeTrade}
          disabled={tradeLoading || !ticker.trim() || !shares}
          className={`px-6 py-2 text-sm font-mono rounded-lg border transition-colors disabled:opacity-40 ${
            action === 'buy'
              ? 'bg-accent/20 border-accent/30 text-accent hover:bg-accent/30'
              : 'bg-danger/20 border-danger/30 text-danger hover:bg-danger/30'
          }`}
        >
          {tradeLoading ? '…' : action === 'buy' ? 'Buy' : 'Sell'}
        </button>
      </div>

      {/* Quote display */}
      {quote && (
        <div className="flex items-center gap-4 mb-3 p-3 rounded-lg bg-white/[0.02] border border-white/[0.04] font-mono text-sm">
          <span className="text-content-muted text-[11px] uppercase tracking-wide">{quote.ticker}</span>
          <span className="text-content-primary font-semibold">{fmt(quote.price)}</span>
          <span className={dirClass(quote.dir)}>
            {quote.change >= 0 ? '+' : ''}{fmt(quote.change)} ({fmtPct(quote.changePct)})
          </span>
          {estimatedCost !== null && (
            <span className="text-content-muted ml-auto">
              Est. {action === 'buy' ? 'cost' : 'proceeds'}: {fmt(estimatedCost)}
            </span>
          )}
        </div>
      )}

      {/* Cash available */}
      <p className="text-[12px] font-mono text-content-muted">
        Available cash: <span className="text-content-secondary">{fmt(cashBalance)}</span>
      </p>

      {/* Feedback */}
      {message && (
        <p className={`mt-3 text-sm font-mono ${message.ok ? 'text-accent' : 'text-danger'}`}>
          {message.ok ? '✓' : '✗'} {message.text}
        </p>
      )}
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const Simulation = () => {
  const { data, loading, error, retry } = useSimulation()
  const [sellPrefill, setSellPrefill] = useState<{ ticker: string; shares: number } | null>(null)
  const [resetting, setResetting] = useState(false)

  const handleSell = useCallback((ticker: string, maxShares: number) => {
    setSellPrefill({ ticker, shares: maxShares })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [])

  const handleReset = useCallback(async () => {
    if (!confirm('Reset simulation? This will clear all positions and restore $100,000 starting cash.')) return
    setResetting(true)
    try {
      await marketService.simulationReset()
      retry()
    } finally {
      setResetting(false)
    }
  }, [retry])

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { cashBalance, startingCash, portfolioValue, totalValue, totalGain, totalGainPct, positions, transactions, performanceData } = data

  const gainDir = totalGain > 0 ? 'up' : totalGain < 0 ? 'down' : 'flat'

  const perfConfig = {
    labels: performanceData.labels,
    datasets: [{
      label: 'Portfolio Value',
      data: performanceData.data,
      borderColor: '#a78bfa',
      backgroundColor: 'rgba(167,139,250,0.06)',
      fill: true, tension: 0.4, pointRadius: 3, pointHoverRadius: 6, borderWidth: 2,
    }],
  }

  const perfOpts = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: axisStyle,
      y: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          callback: (v: number | string) => `$${(Number(v) / 1000).toFixed(0)}k`,
        },
      },
    },
  }

  return (
    <div>
      <div className="flex items-start justify-between mb-6">
        <SectionHeader title="◉ Simulator" subtitle="Virtual Investment Portfolio · $100k Starting Cash" />
        <button
          onClick={handleReset}
          disabled={resetting}
          className="text-[11px] font-mono px-3 py-1.5 rounded border border-white/[0.08] text-content-muted hover:text-danger hover:border-danger/30 transition-colors disabled:opacity-40"
        >
          {resetting ? 'Resetting…' : 'Reset Portfolio'}
        </button>
      </div>

      {/* Summary bar */}
      <div className="bg-surface border border-white/[0.04] rounded-card p-5 mb-6 animate-fadeIn">
        <div className="flex items-start justify-between flex-wrap gap-6">
          <div>
            <p className="font-mono text-[10px] uppercase tracking-[2px] text-content-muted mb-1">Total Value</p>
            <p className="text-[32px] font-bold text-content-primary">{fmt(totalValue)}</p>
            <p className={`text-[14px] font-mono mt-1 ${dirClass(gainDir)}`}>
              {totalGain >= 0 ? '+' : ''}{fmt(totalGain)} ({fmtPct(totalGainPct)}) all-time
            </p>
          </div>
          <div className="flex gap-8">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted mb-1">Cash</p>
              <p className="text-[20px] font-semibold text-content-primary">{fmt(cashBalance)}</p>
              <p className="text-[11px] font-mono text-content-muted">
                {((cashBalance / startingCash) * 100).toFixed(1)}% of portfolio
              </p>
            </div>
            <div>
              <p className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted mb-1">Invested</p>
              <p className="text-[20px] font-semibold text-content-primary">{fmt(portfolioValue)}</p>
              <p className="text-[11px] font-mono text-content-muted">{positions.length} position{positions.length !== 1 ? 's' : ''}</p>
            </div>
            <div>
              <p className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted mb-1">Starting Cash</p>
              <p className="text-[20px] font-semibold text-content-secondary">{fmt(startingCash)}</p>
              <p className="text-[11px] font-mono text-content-muted">virtual capital</p>
            </div>
          </div>
        </div>
      </div>

      {/* Performance chart */}
      {performanceData.labels.length > 0 && (
        <div className="bg-surface border border-white/[0.04] rounded-card p-5 mb-6 animate-fadeIn">
          <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-4">
            Portfolio Performance
          </p>
          <div className="h-48">
            <Line data={perfConfig} options={perfOpts} />
          </div>
        </div>
      )}

      {/* Trade form */}
      <TradePanel
        cashBalance={cashBalance}
        prefillTicker={sellPrefill?.ticker ?? ''}
        prefillAction={sellPrefill ? 'sell' : 'buy'}
        prefillMaxShares={sellPrefill?.shares}
        onTraded={() => { setSellPrefill(null); retry() }}
      />

      <HoldingsTable positions={positions} onSell={handleSell} />
      <TxTable transactions={transactions} />
    </div>
  )
}

export default Simulation
