import { Line, Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, BarElement, Tooltip, Filler,
} from 'chart.js'
import { useOverview } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid4, Grid2, Grid3, Card, ChartCard } from './shared'

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

const Overview = () => {
  const { data, loading, error, retry } = useOverview()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { indices, marketDrivers, sp500ChartData, cryptoChartData, intlChartData } = data

  const sp500Config = {
    labels: sp500ChartData.labels,
    datasets: [{
      label: 'S&P 500', data: sp500ChartData.data,
      borderColor: '#ff3366', backgroundColor: 'rgba(255,51,102,0.05)',
      fill: true, tension: 0.4, pointRadius: 0, pointHoverRadius: 5, borderWidth: 2,
    }],
  }

  const cryptoConfig = {
    labels: cryptoChartData.labels,
    datasets: [{
      label: '24h %', data: cryptoChartData.data,
      backgroundColor: cryptoChartData.colors.map(c => c + '80'),
      borderColor: cryptoChartData.colors,
      borderWidth: 1, borderRadius: 6,
    }],
  }

  const intlConfig = {
    labels: intlChartData.labels,
    datasets: [{
      label: 'Weekly %', data: intlChartData.data,
      backgroundColor: intlChartData.data.map(v => v >= 0 ? 'rgba(0,255,136,0.5)' : 'rgba(255,51,102,0.4)'),
      borderColor: intlChartData.data.map(v => v >= 0 ? '#00ff88' : '#ff3366'),
      borderWidth: 1, borderRadius: 4,
    }],
  }

  return (
    <div>
      <SectionHeader title="Market Overview" subtitle="Live data · March 28, 2026" />

      <Grid4>
        {indices.map(idx => (
          <Card key={idx.ticker}>
            <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">{idx.ticker}</p>
            <p className="text-[28px] font-bold tracking-tight leading-tight mt-1">{idx.value}</p>
            <p className={`font-mono text-[13px] mt-1.5 ${idx.dir === 'down' ? 'text-danger' : 'text-accent'}`}>
              {idx.dir === 'down' ? '▼' : '▲'} {idx.change} ({idx.pct})
            </p>
            <p className={`text-[11px] mt-1.5 ${idx.noteColor === 'danger' ? 'text-danger' : 'text-content-muted'}`}>
              {idx.note}
            </p>
          </Card>
        ))}
      </Grid4>

      <ChartCard
        title="S&P 500 · 2026 YTD Performance"
        badge={<span className="text-danger font-mono text-xs">▼ -8.74% from Jan peak</span>}
      >
        <Line data={sp500Config} options={{ ...baseOpts, scales: { x: { ...axisStyle, grid: { display: false }, ticks: { ...axisStyle.ticks, maxRotation: 45, font: { size: 10 } } }, y: { ...axisStyle, ticks: { ...axisStyle.ticks, callback: (v: number | string) => Number(v).toLocaleString() } } } }} />
      </ChartCard>

      <Grid2 className="mb-0">
        <ChartCard title="Crypto Markets" badge={<span className="text-danger font-mono text-xs">Market cap: $2.37T (-3.1%)</span>} className="mb-0">
          <Bar data={cryptoConfig} options={{ ...baseOpts, scales: { x: { ...axisStyle, grid: { display: false } }, y: { ...axisStyle, ticks: { ...axisStyle.ticks, callback: (v: number | string) => `${v}%` } } } }} />
        </ChartCard>
        <ChartCard title="International Indices" badge={<span className="text-content-secondary font-mono text-xs">Weekly change</span>} className="mb-0">
          <Bar data={intlConfig} options={{ ...baseOpts, scales: { x: { ...axisStyle, grid: { display: false } }, y: { ...axisStyle, ticks: { ...axisStyle.ticks, callback: (v: number | string) => `${v}%` } } } }} />
        </ChartCard>
      </Grid2>

      <div className="mt-6">
        <SectionHeader title="Market Drivers" subtitle="Key catalysts this week" />
        <Grid3>
          {marketDrivers.map(d => (
            <Card key={d.title} style={{ borderLeft: `3px solid var(--tw-${d.color}, #fff)` }}
              className={`border-l-[3px] ${d.color === 'danger' ? 'border-l-danger' : d.color === 'warn' ? 'border-l-warn' : 'border-l-brand-blue'}`}>
              <p className="text-[13px] font-semibold mb-1.5">{d.icon} {d.title}</p>
              <p className="text-[12px] text-content-secondary leading-relaxed">{d.text}</p>
            </Card>
          ))}
        </Grid3>
      </div>
    </div>
  )
}

export default Overview
