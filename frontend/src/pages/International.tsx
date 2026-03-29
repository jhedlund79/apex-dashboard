import { useMarkets } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid2, Card, ChartCard, Table, Th, Td, TickerCell, dirClass } from './shared'

const International = () => {
  const { data, loading, error, retry } = useMarkets()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { intlIndices } = data

  return (
    <div>
      <SectionHeader title="International Markets" subtitle="Global indices" />

      <ChartCard title="Global Index Performance" fixedHeight={false}>
        <div>
          <Table>
            <thead>
              <tr><Th>Index</Th><Th>Region</Th><Th>This Week</Th><Th>Key Driver</Th></tr>
            </thead>
            <tbody>
              {intlIndices.map(row => (
                <tr key={row.ticker} className="hover:[&>td]:bg-white/[0.02]">
                  <TickerCell>{row.ticker}</TickerCell>
                  <Td>{row.region}</Td>
                  <Td className={dirClass(row.dir)}>{row.change}</Td>
                  <Td>{row.driver}</Td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </ChartCard>

      <Grid2>
        <Card className="border-l-[3px] border-l-brand-blue">
          <p className="text-[13px] font-semibold mb-1.5">🇪🇺 Europe Outlook</p>
          <p className="text-[12px] text-content-secondary leading-relaxed">
            ECB raised 2026 inflation forecast to 2.6% from 1.9%. Germany embarking on massive fiscal stimulus.
            European valuations still below US peers despite recent rally. Key risk: prolonged energy shock.
          </p>
        </Card>
        <Card className="border-l-[3px] border-l-brand-cyan">
          <p className="text-[13px] font-semibold mb-1.5">🌏 Asia-Pacific Outlook</p>
          <p className="text-[12px] text-content-secondary leading-relaxed">
            South Korea &amp; Japan led 2025 performance. China AI sector providing growth despite trade tensions.
            EM stocks have outperformed S&amp;P 500 YTD. TSMC a notable contributor to Asian gains.
          </p>
        </Card>
      </Grid2>
    </div>
  )
}

export default International
