import { useRecommendations } from '../hooks/useMarketData'
import { LoadingState } from '../components/LoadingState'
import { ErrorState } from '../components/ErrorState'
import { SectionHeader, Grid3, dirClass } from './shared'
import type { Recommendation } from '../types/market'

const badgeClass: Record<Recommendation['type'], string> = {
  'strong-buy': 'bg-accent/10 text-accent border border-accent/20',
  'buy': 'bg-brand-blue/10 text-brand-blue border border-brand-blue/20',
  'speculative': 'bg-warn/10 text-warn border border-warn/20',
}

const topBorderClass: Record<Recommendation['type'], string> = {
  'strong-buy': 'before:bg-gradient-to-r before:from-accent before:to-brand-cyan',
  'buy': 'before:bg-gradient-to-r before:from-brand-blue before:to-brand-purple',
  'speculative': 'before:bg-gradient-to-r before:from-warn before:to-danger',
}

const RecCard = ({ rec }: { rec: Recommendation }) => (
  <div className={`bg-surface border border-white/[0.06] rounded-card p-5 relative overflow-hidden transition-all duration-300 animate-fadeIn hover:border-accent/20 hover:-translate-y-0.5 before:content-[''] before:absolute before:top-0 before:left-0 before:right-0 before:h-[3px] before:rounded-t-card ${topBorderClass[rec.type]}`}>
    <div className="flex items-start justify-between mb-3">
      <div>
        <p className="font-mono text-[20px] font-bold text-accent">{rec.ticker}</p>
        <p className="text-[12px] text-content-secondary mt-0.5">{rec.name}</p>
      </div>
      <span className={`text-[10px] px-2.5 py-0.5 rounded-full font-mono uppercase tracking-wider ${badgeClass[rec.type]}`}>
        {rec.badge}
      </span>
    </div>
    <div className="grid grid-cols-3 gap-2 mt-3.5 pt-3.5 border-t border-white/[0.04]">
      {rec.stats.map(s => (
        <div key={s.label}>
          <p className="text-[10px] text-content-muted uppercase font-mono tracking-[0.5px]">{s.label}</p>
          <p className={`text-[14px] font-semibold mt-0.5 ${dirClass(s.dir)}`}>{s.value}</p>
        </div>
      ))}
    </div>
    <p className="text-[12px] text-content-secondary leading-relaxed mt-2.5 p-2.5 bg-white/[0.02] rounded-lg">
      {rec.thesis}
    </p>
  </div>
)

const RecGroup = ({ title, items }: { title: string; items: Recommendation[] }) => (
  <>
    <p className="text-[15px] font-semibold mb-3.5">{title}</p>
    <Grid3 className="mb-7">
      {items.map(rec => <RecCard key={rec.ticker} rec={rec} />)}
    </Grid3>
  </>
)

const Recommendations = () => {
  const { data, loading, error, retry } = useRecommendations()

  if (loading) return <LoadingState />
  if (error || !data) return <ErrorState error={error} onRetry={retry} />

  return (
    <div>
      <SectionHeader title="⚡ Aggressive Growth Picks" subtitle="High conviction · High risk" />

      <div className="p-4 bg-accent/[0.04] border border-accent/10 rounded-xl mb-6 text-[12px] text-content-secondary leading-relaxed">
        <strong className="text-accent">Strategy Note:</strong> Current market conditions — extreme fear, Nasdaq
        in correction, elevated VIX — historically present strong entry points for aggressive growth investors
        with 12–24 month horizons. Focus on AI infrastructure leaders, defense beneficiaries, and international
        diversification plays.
      </div>

      <RecGroup title="AI & Semiconductor Growth" items={data.aiSemis} />
      <RecGroup title="High-Growth Disruptors" items={data.disruptors} />
      <RecGroup title="Defense & Infrastructure" items={data.defense} />
      <RecGroup title="Crypto Exposure" items={data.crypto} />

      <div className="mt-8 p-4 bg-warn/[0.05] border border-warn/10 rounded-xl text-[11px] text-content-muted leading-relaxed">
        <strong className="text-warn">⚠ Important Disclaimer:</strong> This dashboard is for informational and
        educational purposes only. It does not constitute financial advice. All investments carry risk. Always
        consult a qualified financial advisor before making investment decisions.
      </div>
    </div>
  )
}

export default Recommendations
