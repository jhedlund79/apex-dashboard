import { useMarkets } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid3, Card, ChartCard, Table, Th, Td, TickerCell, dirClass } from './shared'

const USMarkets = () => {
  const { data, loading, error, retry } = useMarkets()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  const { usSummary, usMovers } = data

  return (
    <div>
      <SectionHeader title="US Markets" subtitle="Major indices & movers" />

      <Grid3>
        <Card>
          <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">S&amp;P 500 March</p>
          <p className="text-[24px] font-bold text-danger mt-1">{usSummary.marchReturn}</p>
          <p className="text-[11px] text-content-muted mt-1.5">{usSummary.marchNote}</p>
        </Card>
        <Card>
          <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">S&amp;P 500 YoY</p>
          <p className="text-[24px] font-bold text-accent mt-1">{usSummary.yoYReturn}</p>
          <p className="text-[11px] text-content-muted mt-1.5">{usSummary.yoYNote}</p>
        </Card>
        <Card>
          <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">Consumer Sentiment</p>
          <p className="text-[24px] font-bold mt-1">{usSummary.consumerSentiment}</p>
          <p className="text-[11px] text-danger mt-1.5">{usSummary.sentimentNote}</p>
        </Card>
      </Grid3>

      <ChartCard title="Notable US Movers This Week" fixedHeight={false}>
        <div>
          <Table>
            <thead>
              <tr><Th>Ticker</Th><Th>Name</Th><Th>Move</Th><Th>Driver</Th></tr>
            </thead>
            <tbody>
              {usMovers.map(row => (
                <tr key={row.ticker} className="hover:[&>td]:bg-white/[0.02]">
                  <TickerCell>{row.ticker}</TickerCell>
                  <Td>{row.name}</Td>
                  <Td className={dirClass(row.dir)}>{row.move}</Td>
                  <Td>{row.driver}</Td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </ChartCard>
    </div>
  )
}

export default USMarkets
