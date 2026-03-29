import { Bar } from 'react-chartjs-2'
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Tooltip, Legend } from 'chart.js'
import { useMarkets } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid2, Card, ChartCard } from './shared'

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip, Legend)

const axisStyle = {
  grid: { color: 'rgba(255,255,255,0.03)' },
  ticks: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } },
}

const Sectors = () => {
  const { data, loading, error, retry } = useMarkets()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { sectorData } = data

  const chartConfig = {
    labels: sectorData.labels,
    datasets: [
      { label: 'Current Momentum', data: sectorData.momentum, backgroundColor: 'rgba(0,255,136,0.4)', borderColor: '#00ff88', borderWidth: 1, borderRadius: 4 },
      { label: '6M Outlook Score', data: sectorData.outlook, backgroundColor: 'rgba(68,136,255,0.4)', borderColor: '#4488ff', borderWidth: 1, borderRadius: 4 },
    ],
  }

  return (
    <div>
      <SectionHeader title="Sector Analysis" subtitle="Opportunity heat map" />

      <ChartCard title="Sector Performance & Outlook" className="mb-6">
        <div className="relative h-[320px]">
          <Bar data={chartConfig} options={{
            responsive: true, maintainAspectRatio: false, indexAxis: 'y',
            plugins: { legend: { labels: { boxWidth: 12, padding: 20, color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } } } },
            scales: { x: { ...axisStyle, max: 100 }, y: { ...axisStyle, grid: { display: false } } },
          }} />
        </div>
      </ChartCard>

      <Grid2>
        <Card className="border-l-[3px] border-l-accent">
          <p className="text-[14px] font-semibold mb-2 text-accent">Sectors to Overweight</p>
          <p className="text-[10px] text-content-muted font-mono uppercase tracking-wider mb-2.5">Ranked by combined 1M + 3M momentum</p>
          <div className="text-[12px] text-content-secondary leading-[1.8]">
            {sectorData.overweight.map((s, i) => (
              <div key={s} className="flex items-center gap-2">
                <span className="text-accent font-mono text-[10px] w-4">{i + 1}.</span>
                <span className="font-semibold">{s}</span>
              </div>
            ))}
          </div>
        </Card>
        <Card className="border-l-[3px] border-l-danger">
          <p className="text-[14px] font-semibold mb-2 text-danger">Sectors to Underweight</p>
          <p className="text-[10px] text-content-muted font-mono uppercase tracking-wider mb-2.5">Weakest by combined 1M + 3M momentum</p>
          <div className="text-[12px] text-content-secondary leading-[1.8]">
            {sectorData.underweight.map((s, i) => (
              <div key={s} className="flex items-center gap-2">
                <span className="text-danger font-mono text-[10px] w-4">{i + 1}.</span>
                <span className="font-semibold">{s}</span>
              </div>
            ))}
          </div>
        </Card>
      </Grid2>
    </div>
  )
}

export default Sectors
