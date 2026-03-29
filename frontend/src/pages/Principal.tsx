import { Line, Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, BarElement, Tooltip, Filler,
} from 'chart.js'
import { useCallback } from 'react'
import { usePrincipal } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { PlaidConnect } from '../components/PlaidConnect'
import {
  SectionHeader, Grid2, Grid4, Card, ChartCard,
  Table, Th, Td, TickerCell, dirClass,
} from './shared'
import type { PortfolioSummary, PortfolioHolding, Contribution, ChangeMetric } from '../types/market'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, Tooltip, Filler)

const axisStyle = {
  grid: { color: 'rgba(255,255,255,0.03)' },
  ticks: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } },
}

const baseOpts = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
} as const

const dirCls = (dir: string) =>
  dir === 'up' ? 'text-accent' : dir === 'down' ? 'text-danger' : 'text-content-primary'

const ChangeCard = ({ label, metric }: { label: string; metric: ChangeMetric }) => (
  <Card>
    <p className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted mb-2">{label}</p>
    <p className={`text-[22px] font-bold ${dirCls(metric.dir)}`}>{metric.amount}</p>
    <p className={`text-[12px] font-mono mt-0.5 ${dirCls(metric.dir)}`}>{metric.pct}</p>
  </Card>
)

const SummaryBar = ({ s }: { s: PortfolioSummary }) => (
  <div className="bg-surface border border-white/[0.04] rounded-card p-5 mb-6 animate-fadeIn">
    <div className="flex items-start justify-between flex-wrap gap-4">
      <div>
        <p className="font-mono text-[10px] uppercase tracking-[2px] text-content-muted mb-1">{s.provider} · {s.accountType}</p>
        <p className="text-[28px] font-bold text-content-primary">{s.currentValue}</p>
        <p className="text-[13px] text-content-secondary mt-0.5">{s.accountName}</p>
      </div>
      <div className="text-right">
        <p className="text-[11px] text-content-muted mb-1">Total gain / cost basis</p>
        <p className={`text-[20px] font-semibold ${dirCls(s.totalGainDir)}`}>{s.totalGain}</p>
        <p className={`text-[13px] font-mono ${dirCls(s.totalGainDir)}`}>
          {s.totalGainPct} · cost {s.totalCost}
        </p>
      </div>
    </div>
  </div>
)

const HoldingsTable = ({ holdings }: { holdings: PortfolioHolding[] }) => (
  <div className="bg-surface border border-white/[0.04] rounded-card overflow-hidden mb-6 animate-fadeIn">
    <div className="p-5 pb-0">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-3.5">Holdings</p>
    </div>
    <Table>
      <thead>
        <tr>
          <Th>Fund</Th>
          <Th>Name</Th>
          <Th>Value</Th>
          <Th>Allocation</Th>
          <Th>Gain / Loss</Th>
        </tr>
      </thead>
      <tbody>
        {holdings.map(h => (
          <tr key={h.ticker} className="hover:bg-white/[0.02] transition-colors">
            <TickerCell>{h.ticker}</TickerCell>
            <Td className="text-content-secondary">{h.name}</Td>
            <Td className="font-mono">{h.value}</Td>
            <Td>
              <div className="flex items-center gap-2">
                <div className="h-1.5 bg-accent/20 rounded-full w-20">
                  <div
                    className="h-full bg-accent rounded-full"
                    style={{ width: h.allocation }}
                  />
                </div>
                <span className="font-mono text-[12px]">{h.allocation}</span>
              </div>
            </Td>
            <Td className={`font-mono ${dirClass(h.gainDir)}`}>{h.gain}</Td>
          </tr>
        ))}
      </tbody>
    </Table>
  </div>
)

const ContribTypeLabel: Record<Contribution['type'], string> = {
  employee: 'Employee + Employer',
  employer: 'Employer',
  rollover: 'Rollover',
  deposit: 'Deposit',
}

const ContribTable = ({ contributions }: { contributions: Contribution[] }) => (
  <div className="bg-surface border border-white/[0.04] rounded-card overflow-hidden mb-6 animate-fadeIn">
    <div className="p-5 pb-0">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-3.5">Contribution History</p>
    </div>
    <Table>
      <thead>
        <tr>
          <Th>Date</Th>
          <Th>Amount</Th>
          <Th>Type</Th>
          <Th>Note</Th>
        </tr>
      </thead>
      <tbody>
        {contributions.map((c, i) => (
          <tr key={i} className="hover:bg-white/[0.02] transition-colors">
            <Td className="font-mono text-content-muted">{c.date}</Td>
            <Td className="font-mono text-accent font-semibold">{c.amount}</Td>
            <Td>
              <span className="text-[11px] px-2 py-0.5 rounded-full bg-accent/10 text-accent border border-accent/20 font-mono uppercase tracking-wide">
                {ContribTypeLabel[c.type]}
              </span>
            </Td>
            <Td className="text-content-secondary">{c.note}</Td>
          </tr>
        ))}
      </tbody>
    </Table>
  </div>
)

const Principal = () => {
  const { data, loading, error, retry } = usePrincipal()

  const handleConnected = useCallback(() => retry(), [retry])

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />
  if (!data.connected) {
    return (
      <div>
        <SectionHeader title="◎ Principal Financial Group" subtitle="401(k) Retirement Account" />
        <PlaidConnect slot="principal" institutionName="Principal Financial Group" onConnected={handleConnected} />
      </div>
    )
  }

  const { summary, performanceData, contribData, holdings, contributions } = data

  const perfConfig = {
    labels: performanceData.labels,
    datasets: [{
      label: 'Balance',
      data: performanceData.data,
      borderColor: '#00ff88',
      backgroundColor: 'rgba(0,255,136,0.06)',
      fill: true, tension: 0.4, pointRadius: 3, pointHoverRadius: 6, borderWidth: 2,
    }],
  }

  const contribConfig = {
    labels: contribData.labels,
    datasets: [{
      label: 'Contributions',
      data: contribData.data,
      backgroundColor: 'rgba(0,255,136,0.4)',
      borderColor: '#00ff88',
      borderWidth: 1, borderRadius: 4,
    }],
  }

  const perfOpts = {
    ...baseOpts,
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

  const contribOpts = {
    ...baseOpts,
    scales: {
      x: axisStyle,
      y: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          callback: (v: number | string) => `$${Number(v).toLocaleString()}`,
        },
      },
    },
  }

  return (
    <div>
      <SectionHeader title="◎ Principal Financial Group" subtitle="401(k) Retirement Account" />

      <SummaryBar s={summary} />

      <Grid4 className="mb-6">
        <ChangeCard label="Weekly Change" metric={summary.weeklyChange} />
        <ChangeCard label="Monthly Change" metric={summary.monthlyChange} />
        <ChangeCard label="YTD Change" metric={summary.ytdChange} />
        <Card>
          <p className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted mb-2">Total Return</p>
          <p className={`text-[22px] font-bold ${dirCls(summary.totalGainDir)}`}>{summary.totalGainPct}</p>
          <p className="text-[12px] font-mono text-content-secondary mt-0.5">{summary.totalGain} all-time</p>
        </Card>
      </Grid4>

      <Grid2 className="mb-6">
        <ChartCard title="Portfolio Balance — 12 Month">
          <Line data={perfConfig} options={perfOpts} />
        </ChartCard>
        <ChartCard title="Monthly Contributions">
          <Bar data={contribConfig} options={contribOpts} />
        </ChartCard>
      </Grid2>

      <HoldingsTable holdings={holdings} />
      <ContribTable contributions={contributions} />
    </div>
  )
}

export default Principal
