import { Line } from 'react-chartjs-2'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler } from 'chart.js'
import { useCrypto } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid4, Grid3, Card, ChartCard, dirClass } from './shared'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const axisStyle = {
  grid: { color: 'rgba(255,255,255,0.03)' },
  ticks: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } },
}

const Crypto = () => {
  const { data, loading, error, retry } = useCrypto()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { cryptoAssets, cryptoSmall, btcHistoryData, marketContext } = data

  const btcConfig = {
    labels: btcHistoryData.labels,
    datasets: [{
      label: 'BTC Price', data: btcHistoryData.data,
      borderColor: '#f7931a', backgroundColor: 'rgba(247,147,26,0.05)',
      fill: true, tension: 0.3, pointRadius: 3, pointHoverRadius: 6, borderWidth: 2,
      pointBackgroundColor: '#f7931a',
    }],
  }

  return (
    <div>
      <SectionHeader title="Crypto Markets" subtitle="Fear & Greed: 12 — Extreme Fear" />

      <Grid4>
        {cryptoAssets.map(asset => (
          <Card key={asset.ticker}>
            <div className="flex items-center gap-2.5 mb-2">
              <div
                className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold font-mono shrink-0"
                style={{ background: asset.bg, color: asset.color }}
              >
                {asset.symbol}
              </div>
              <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">{asset.name}</p>
            </div>
            <p className="text-[22px] font-bold">{asset.price}</p>
            <p className={`font-mono text-[13px] mt-1 ${dirClass(asset.dir)}`}>
              {asset.dir === 'down' ? '▼' : '▲'} {asset.change} (24h)
            </p>
            {asset.meta && <p className="text-[11px] text-content-muted mt-1">{asset.meta}</p>}
          </Card>
        ))}
      </Grid4>

      <Grid3>
        {cryptoSmall.map(c => (
          <Card key={c.ticker}>
            <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">{c.ticker}</p>
            <p className="text-[20px] font-bold mt-1">{c.price}</p>
            <p className={`font-mono text-[13px] mt-1 ${dirClass(c.dir)}`}>
              {c.dir === 'down' ? '▼' : '▲'} {c.change}
            </p>
          </Card>
        ))}
      </Grid3>

      <ChartCard title="Bitcoin · 12 Month Price History">
        <Line data={btcConfig} options={{
          responsive: true, maintainAspectRatio: false,
          plugins: { legend: { display: false } },
          scales: {
            x: { ...axisStyle, grid: { display: false } },
            y: { ...axisStyle, ticks: { ...axisStyle.ticks, callback: (v: number | string) => `$${(Number(v) / 1000).toFixed(0)}K` } },
          },
        }} />
      </ChartCard>

      <Card className="border-l-[3px] border-l-warn">
        <p className="text-[13px] font-semibold mb-1.5">Crypto Market Context</p>
        <p className="text-[12px] text-content-secondary leading-relaxed">{marketContext}</p>
      </Card>
    </div>
  )
}

export default Crypto
