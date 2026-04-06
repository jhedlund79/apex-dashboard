import { Line, Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Tooltip,
  Filler,
  Legend,
} from 'chart.js'
import { useState, useMemo } from 'react'
import { SectionHeader, ChartCard, Card, Grid2 } from './shared'
import {
  computeScenario,
  findBreakEvenAge,
  findDepletionAge,
  type RetirementInputs as Inputs,
  type YearData,
} from '../utils/retirementCalc'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, Tooltip, Filler, Legend)

// ── Formatters ────────────────────────────────────────────────────────────────

const fmt = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(n)

const fmtK = (n: number) => {
  if (n >= 1_000_000) return `$${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `$${(n / 1_000).toFixed(0)}K`
  return fmt(n)
}

const fmtPct = (n: number) => `${(n * 100).toFixed(1)}%`

// ── Calculation Engine ────────────────────────────────────────────────────────

// ── Chart Styles ──────────────────────────────────────────────────────────────

const axisStyle = {
  grid: { color: 'rgba(255,255,255,0.03)' },
  ticks: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 } },
}

const AMBER = '#ffaa00'
const GREEN = '#00ff88'
const BLUE  = '#4488ff'

// ── Input Controls ────────────────────────────────────────────────────────────

interface SliderProps {
  label: string
  value: number
  min: number
  max: number
  step: number
  format: (v: number) => string
  onChange: (v: number) => void
}

const Slider = ({ label, value, min, max, step, format, onChange }: SliderProps) => (
  <div>
    <div className="flex justify-between items-baseline mb-1">
      <span className="text-[11px] font-mono uppercase tracking-[1.5px] text-content-muted">{label}</span>
      <span className="text-sm font-mono font-semibold text-content-primary">{format(value)}</span>
    </div>
    <input
      type="range"
      min={min}
      max={max}
      step={step}
      value={value}
      onChange={e => onChange(Number(e.target.value))}
      className="w-full h-1 rounded appearance-none bg-white/10 cursor-pointer"
      style={{ accentColor: GREEN }}
    />
  </div>
)

// ── Scenario Summary Card ─────────────────────────────────────────────────────

interface ScenarioMeta {
  retirementAge: 62 | 67
  balanceAtRetirement: number
  annualSS: number
  annualWithdrawal: number
  totalAnnualIncome: number
  depletionAge: number | null
  color: string
}

const ScenarioCard = ({ meta }: { meta: ScenarioMeta }) => {
  const { retirementAge, balanceAtRetirement, annualSS, annualWithdrawal, totalAnnualIncome, depletionAge, color } = meta
  return (
    <div
      className="bg-surface border rounded-card p-5 animate-fadeIn"
      style={{ borderColor: `${color}22` }}
    >
      <div className="flex items-center gap-2 mb-4">
        <div className="w-2 h-2 rounded-full" style={{ backgroundColor: color }} />
        <p className="font-mono text-[11px] uppercase tracking-[2px]" style={{ color }}>
          Retire at {retirementAge}
          {retirementAge === 62 ? ' · Early' : ' · Full Retirement Age'}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <p className="text-[10px] font-mono uppercase tracking-[1.5px] text-content-muted mb-1">
            401K at Retirement
          </p>
          <p className="text-[20px] font-bold text-content-primary">{fmtK(balanceAtRetirement)}</p>
        </div>
        <div>
          <p className="text-[10px] font-mono uppercase tracking-[1.5px] text-content-muted mb-1">
            Annual SS Benefit
          </p>
          <p className="text-[20px] font-bold text-content-primary">{fmtK(annualSS)}</p>
          {retirementAge === 62 && (
            <p className="text-[10px] font-mono text-content-muted">30% reduction applied</p>
          )}
        </div>
        <div>
          <p className="text-[10px] font-mono uppercase tracking-[1.5px] text-content-muted mb-1">
            Annual Withdrawal (4%)
          </p>
          <p className="text-[20px] font-bold text-content-primary">{fmtK(annualWithdrawal)}</p>
        </div>
        <div>
          <p className="text-[10px] font-mono uppercase tracking-[1.5px] text-content-muted mb-1">
            Total Annual Income
          </p>
          <p className="text-[20px] font-bold" style={{ color }}>{fmtK(totalAnnualIncome)}</p>
        </div>
      </div>

      {depletionAge && (
        <div className="mt-4 pt-4 border-t border-white/[0.04]">
          <p className="text-[11px] font-mono text-danger">
            ⚠ 401K depletes at age {depletionAge}. SS income continues.
          </p>
        </div>
      )}
      {!depletionAge && (
        <div className="mt-4 pt-4 border-t border-white/[0.04]">
          <p className="text-[11px] font-mono text-content-muted">
            ✓ Balance sustained through life expectancy
          </p>
        </div>
      )}
    </div>
  )
}

// ── Break-Even Card ───────────────────────────────────────────────────────────

const BreakEvenCard = ({ breakEvenAge, lifeExpectancy }: { breakEvenAge: number | null; lifeExpectancy: number }) => (
  <div className="bg-surface border border-white/[0.04] rounded-card p-5 animate-fadeIn mb-6">
    <p className="font-mono text-[11px] uppercase tracking-[2px] text-content-muted mb-3">
      Break-Even Analysis
    </p>
    {breakEvenAge ? (
      <div className="flex items-center gap-6">
        <div>
          <p className="text-[32px] font-bold text-content-primary">Age {breakEvenAge}</p>
          <p className="text-sm text-content-secondary mt-1">
            Retiring at 67 surpasses retiring at 62 in cumulative lifetime income
          </p>
        </div>
        <div className="flex-1 h-px bg-white/[0.04] hidden lg:block" />
        <div className="text-right">
          <p className="text-[12px] font-mono text-content-muted">Years after retiring early</p>
          <p className="text-[20px] font-semibold" style={{ color: GREEN }}>{breakEvenAge - 62} yrs</p>
          <p className="text-[12px] font-mono text-content-muted">Years of retirement ahead</p>
          <p className="text-[20px] font-semibold text-content-primary">{lifeExpectancy - breakEvenAge} yrs</p>
        </div>
      </div>
    ) : (
      <p className="text-lg text-warn">
        Retiring at 62 produces more cumulative income through age {lifeExpectancy}. Consider a longer time horizon.
      </p>
    )}
  </div>
)

// ── Wealth Trajectory Chart ───────────────────────────────────────────────────

const WealthChart = ({
  s62,
  s67,
  currentAge,
}: {
  s62: YearData[]
  s67: YearData[]
  currentAge: number
}) => {
  const allAges = s62.map(d => d.age)
  const labels = allAges.map(a => String(a))
  const map67 = new Map(s67.map(d => [d.age, d]))

  const data = {
    labels,
    datasets: [
      {
        label: 'Retire @ 62',
        data: s62.map(d => d.balance),
        borderColor: AMBER,
        backgroundColor: 'rgba(255,170,0,0.05)',
        fill: false,
        tension: 0.3,
        pointRadius: 0,
        pointHoverRadius: 4,
        borderWidth: 2,
      },
      {
        label: 'Retire @ 67',
        data: allAges.map(a => map67.get(a)?.balance ?? null),
        borderColor: GREEN,
        backgroundColor: 'rgba(0,255,136,0.05)',
        fill: false,
        tension: 0.3,
        pointRadius: 0,
        pointHoverRadius: 4,
        borderWidth: 2,
      },
    ],
  }

  const opts = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      legend: {
        display: true,
        labels: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 }, boxWidth: 12 },
      },
      tooltip: {
        callbacks: {
          title: (items: { label: string }[]) => `Age ${items[0]?.label}`,
          label: (item: { dataset: { label?: string }; raw: unknown }) =>
            ` ${item.dataset.label}: ${fmtK(Number(item.raw))}`,
        },
      },
    },
    scales: {
      x: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          maxTicksLimit: 12,
          callback: (_: unknown, i: number) => (i % 5 === 0 ? String(currentAge + i) : ''),
        },
      },
      y: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          callback: (v: number | string) => fmtK(Number(v)),
        },
      },
    },
  }

  return <Line data={data} options={opts} />
}

// ── Income Streams Chart ──────────────────────────────────────────────────────

const IncomeStreamsChart = ({
  rows,
  color,
  retirementAge,
}: {
  rows: YearData[]
  color: string
  retirementAge: 62 | 67
}) => {
  const retRows = rows.filter(d => d.phase === 'retirement')
  const labels = retRows.map(d => String(d.age))

  const ssColor = color === AMBER ? 'rgba(255,170,0,0.7)' : 'rgba(0,255,136,0.7)'
  const wdColor = color === AMBER ? 'rgba(68,136,255,0.7)' : 'rgba(136,85,255,0.7)'

  const data = {
    labels,
    datasets: [
      {
        label: 'Social Security',
        data: retRows.map(d => Math.round(d.ssIncome)),
        backgroundColor: ssColor,
        borderRadius: 2,
        stack: 'income',
      },
      {
        label: '401K Withdrawal',
        data: retRows.map(d => Math.round(d.withdrawal)),
        backgroundColor: wdColor,
        borderRadius: 2,
        stack: 'income',
      },
    ],
  }

  const opts = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      legend: {
        display: true,
        labels: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 10 }, boxWidth: 10 },
      },
      tooltip: {
        callbacks: {
          title: (items: { label: string }[]) => `Age ${items[0]?.label}`,
          label: (item: { dataset: { label?: string }; raw: unknown }) =>
            ` ${item.dataset.label}: ${fmtK(Number(item.raw))}`,
        },
      },
    },
    scales: {
      x: {
        ...axisStyle,
        stacked: true,
        ticks: {
          ...axisStyle.ticks,
          maxTicksLimit: 10,
          callback: (_: unknown, i: number) => (i % 5 === 0 ? String(retirementAge + i) : ''),
        },
      },
      y: {
        ...axisStyle,
        stacked: true,
        ticks: {
          ...axisStyle.ticks,
          callback: (v: number | string) => fmtK(Number(v)),
        },
      },
    },
  }

  return <Bar data={data} options={opts} />
}

// ── Break-Even Chart ──────────────────────────────────────────────────────────

const BreakEvenChart = ({
  s62,
  s67,
  breakEvenAge,
}: {
  s62: YearData[]
  s67: YearData[]
  breakEvenAge: number | null
}) => {
  const retRows62 = s62.filter(d => d.phase === 'retirement')
  const map67 = new Map(s67.filter(d => d.phase === 'retirement').map(d => [d.age, d]))
  const labels = retRows62.map(d => String(d.age))

  const data = {
    labels,
    datasets: [
      {
        label: 'Cumulative — Retire @ 62',
        data: retRows62.map(d => Math.round(d.cumulativeRetirementIncome)),
        borderColor: AMBER,
        backgroundColor: 'rgba(255,170,0,0.05)',
        fill: false,
        tension: 0.3,
        pointRadius: 0,
        pointHoverRadius: 4,
        borderWidth: 2,
      },
      {
        label: 'Cumulative — Retire @ 67',
        data: retRows62.map(d => {
          const d67 = map67.get(d.age)
          return d67 ? Math.round(d67.cumulativeRetirementIncome) : 0
        }),
        borderColor: GREEN,
        backgroundColor: 'rgba(0,255,136,0.05)',
        fill: false,
        tension: 0.3,
        pointRadius: 0,
        pointHoverRadius: 4,
        borderWidth: 2,
      },
      ...(breakEvenAge
        ? [
            {
              label: `Break-even @ ${breakEvenAge}`,
              data: retRows62.map(d => (d.age === breakEvenAge ? map67.get(d.age)?.cumulativeRetirementIncome ?? null : null)),
              borderColor: '#ff3366',
              backgroundColor: '#ff3366',
              pointRadius: 6,
              pointHoverRadius: 8,
              showLine: false,
              borderWidth: 0,
            },
          ]
        : []),
    ],
  }

  const opts = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      legend: {
        display: true,
        labels: { color: '#8888a0', font: { family: "'DM Mono', monospace", size: 11 }, boxWidth: 12 },
      },
      tooltip: {
        callbacks: {
          title: (items: { label: string }[]) => `Age ${items[0]?.label}`,
          label: (item: { dataset: { label?: string }; raw: unknown }) =>
            ` ${item.dataset.label}: ${fmtK(Number(item.raw))}`,
        },
      },
    },
    scales: {
      x: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          maxTicksLimit: 12,
          callback: (_: unknown, i: number) => (i % 5 === 0 ? String(62 + i) : ''),
        },
      },
      y: {
        ...axisStyle,
        ticks: {
          ...axisStyle.ticks,
          callback: (v: number | string) => fmtK(Number(v)),
        },
      },
    },
  }

  return <Line data={data} options={opts} />
}

// ── Projection Table ──────────────────────────────────────────────────────────

const ProjectionTable = ({
  s62,
  s67,
  currentAge,
}: {
  s62: YearData[]
  s67: YearData[]
  currentAge: number
}) => {
  const map67 = new Map(s67.map(d => [d.age, d]))

  // Show every 5 years during accumulation, every year during retirement
  const rows = s62.filter(d => {
    if (d.phase === 'accumulation') return (d.age - currentAge) % 5 === 0
    return true
  })

  return (
    <div className="bg-surface border border-white/[0.04] rounded-card overflow-hidden mb-6 animate-fadeIn">
      <div className="p-5 pb-0">
        <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted mb-3.5">
          Year-by-Year Projection
        </p>
        <p className="text-[11px] text-content-muted mb-4">
          All figures in nominal (future) dollars · 4% annual withdrawal rule in retirement
        </p>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full border-collapse min-w-[760px]">
          <thead>
            <tr>
              <th className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted text-left px-3.5 py-2.5 border-b border-white/[0.06]">
                Age
              </th>
              <th className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted text-left px-3.5 py-2.5 border-b border-white/[0.06]">
                Phase
              </th>
              <th
                className="font-mono text-[10px] uppercase tracking-[1.5px] text-left px-3.5 py-2.5 border-b border-white/[0.06]"
                style={{ color: AMBER }}
                colSpan={3}
              >
                ← Retire @ 62
              </th>
              <th
                className="font-mono text-[10px] uppercase tracking-[1.5px] text-left px-3.5 py-2.5 border-b border-white/[0.06]"
                style={{ color: GREEN }}
                colSpan={3}
              >
                ← Retire @ 67
              </th>
            </tr>
            <tr>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]" />
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]" />
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">Balance</th>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">SS</th>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">Withdrawal</th>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">Balance</th>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">SS</th>
              <th className="font-mono text-[10px] text-content-muted text-left px-3.5 pb-2 border-b border-white/[0.06]">Withdrawal</th>
            </tr>
          </thead>
          <tbody>
            {rows.map(d => {
              const d67 = map67.get(d.age)
              const isRetirement62 = d.phase === 'retirement'
              const isRetirement67 = d67?.phase === 'retirement'
              return (
                <tr key={d.age} className="hover:bg-white/[0.02] transition-colors">
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono font-semibold text-[13px] text-content-primary">
                    {d.age}
                  </td>
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] text-[11px] font-mono">
                    <span
                      className={`px-2 py-0.5 rounded-full border ${
                        isRetirement62
                          ? 'border-warn/20 text-warn bg-warn/5'
                          : 'border-white/10 text-content-muted bg-white/[0.02]'
                      }`}
                    >
                      {isRetirement62 ? 'Retired' : 'Working'}
                    </span>
                  </td>
                  {/* Retire @ 62 */}
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px] text-content-primary">
                    {fmtK(d.balance)}
                    {d.balance === 0 && isRetirement62 && (
                      <span className="ml-1 text-[10px] text-danger">depleted</span>
                    )}
                  </td>
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px]" style={{ color: isRetirement62 ? AMBER : '#555570' }}>
                    {isRetirement62 ? fmtK(d.ssIncome) : '—'}
                  </td>
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px]" style={{ color: isRetirement62 ? BLUE : '#555570' }}>
                    {isRetirement62 ? fmtK(d.withdrawal) : '—'}
                  </td>
                  {/* Retire @ 67 */}
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px] text-content-primary">
                    {d67 ? fmtK(d67.balance) : '—'}
                    {d67?.balance === 0 && isRetirement67 && (
                      <span className="ml-1 text-[10px] text-danger">depleted</span>
                    )}
                  </td>
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px]" style={{ color: isRetirement67 ? GREEN : '#555570' }}>
                    {isRetirement67 ? fmtK(d67!.ssIncome) : '—'}
                  </td>
                  <td className="px-3.5 py-2.5 border-b border-white/[0.03] font-mono text-[13px]" style={{ color: isRetirement67 ? '#8855ff' : '#555570' }}>
                    {isRetirement67 ? fmtK(d67!.withdrawal) : '—'}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const DEFAULT_INPUTS: Inputs = {
  currentAge: 45,
  currentBalance: 150_000,
  annualIncome: 100_000,
  contributionRate: 0.10,
  returnRate: 0.07,
  inflationRate: 0.03,
  socialSecurityFRA: 24_000,
  lifeExpectancy: 90,
}

const Retirement = () => {
  const [inputs, setInputs] = useState<Inputs>(DEFAULT_INPUTS)

  const set = (key: keyof Inputs) => (value: number) =>
    setInputs(prev => ({ ...prev, [key]: value }))

  const { s62, s67, meta62, meta67, breakEvenAge } = useMemo(() => {
    const scenario62 = computeScenario(inputs, 62)
    const scenario67 = computeScenario(inputs, 67)

    const ret62 = scenario62.find(d => d.age === 62)!
    const ret67 = scenario67.find(d => d.age === 67)!

    const balAt62 = scenario62.filter(d => d.age < 62).at(-1)?.balance ?? inputs.currentBalance
    const balAt67 = scenario67.filter(d => d.age < 67).at(-1)?.balance ?? inputs.currentBalance

    const initialWithdrawal62 = balAt62 * 0.04
    const initialWithdrawal67 = balAt67 * 0.04

    const meta62: ScenarioMeta = {
      retirementAge: 62,
      balanceAtRetirement: balAt62,
      annualSS: inputs.socialSecurityFRA * 0.7,
      annualWithdrawal: initialWithdrawal62,
      totalAnnualIncome: initialWithdrawal62 + inputs.socialSecurityFRA * 0.7,
      depletionAge: findDepletionAge(scenario62),
      color: AMBER,
    }

    const meta67: ScenarioMeta = {
      retirementAge: 67,
      balanceAtRetirement: balAt67,
      annualSS: inputs.socialSecurityFRA,
      annualWithdrawal: initialWithdrawal67,
      totalAnnualIncome: initialWithdrawal67 + inputs.socialSecurityFRA,
      depletionAge: findDepletionAge(scenario67),
      color: GREEN,
    }

    // Suppress unused-variable warnings
    void ret62
    void ret67

    return {
      s62: scenario62,
      s67: scenario67,
      meta62,
      meta67,
      breakEvenAge: findBreakEvenAge(scenario62, scenario67, inputs.lifeExpectancy),
    }
  }, [inputs])

  return (
    <div>
      <SectionHeader
        title="◎ Retirement Calculator"
        subtitle="62 vs 67 · Retire Early or Wait for Full Benefits"
      />

      {/* ── Inputs ── */}
      <div className="bg-surface border border-white/[0.04] rounded-card p-5 mb-6 animate-fadeIn">
        <p className="font-mono text-[11px] uppercase tracking-[2px] text-content-muted mb-4">
          Parameters
        </p>
        <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
          {/* Left column */}
          <div className="space-y-5">
            <Slider
              label="Current Age"
              value={inputs.currentAge}
              min={25}
              max={61}
              step={1}
              format={v => `${v} yrs`}
              onChange={set('currentAge')}
            />
            <Slider
              label="Current 401K Balance"
              value={inputs.currentBalance}
              min={0}
              max={1_000_000}
              step={5_000}
              format={fmtK}
              onChange={set('currentBalance')}
            />
            <Slider
              label="Annual Income"
              value={inputs.annualIncome}
              min={30_000}
              max={500_000}
              step={5_000}
              format={fmtK}
              onChange={set('annualIncome')}
            />
            <Slider
              label="401K Contribution Rate"
              value={inputs.contributionRate}
              min={0.01}
              max={0.23}
              step={0.01}
              format={fmtPct}
              onChange={set('contributionRate')}
            />
          </div>
          {/* Right column */}
          <div className="space-y-5">
            <Slider
              label="Expected Annual Return"
              value={inputs.returnRate}
              min={0.03}
              max={0.12}
              step={0.005}
              format={fmtPct}
              onChange={set('returnRate')}
            />
            <Slider
              label="Inflation Rate"
              value={inputs.inflationRate}
              min={0.01}
              max={0.08}
              step={0.005}
              format={fmtPct}
              onChange={set('inflationRate')}
            />
            <Slider
              label="Social Security at FRA (age 67)"
              value={inputs.socialSecurityFRA}
              min={6_000}
              max={60_000}
              step={500}
              format={v => `${fmtK(v)} / yr`}
              onChange={set('socialSecurityFRA')}
            />
            <Slider
              label="Life Expectancy"
              value={inputs.lifeExpectancy}
              min={70}
              max={95}
              step={1}
              format={v => `Age ${v}`}
              onChange={set('lifeExpectancy')}
            />
          </div>
        </div>
      </div>

      {/* ── Scenario Summary Cards ── */}
      <Grid2 className="mb-6">
        <ScenarioCard meta={meta62} />
        <ScenarioCard meta={meta67} />
      </Grid2>

      {/* ── Break-Even Card ── */}
      <BreakEvenCard breakEvenAge={breakEvenAge} lifeExpectancy={inputs.lifeExpectancy} />

      {/* ── Wealth Trajectory ── */}
      <ChartCard title="401K Wealth Trajectory · Both Scenarios" className="mb-6" fixedHeight={false}>
        <div className="h-[280px]">
          <WealthChart s62={s62} s67={s67} currentAge={inputs.currentAge} />
        </div>
      </ChartCard>

      {/* ── Income Streams ── */}
      <div className="grid grid-cols-1 gap-4 mb-6 lg:grid-cols-2">
        <ChartCard title={`Income Streams · Retire @ 62`} fixedHeight={false}>
          <div className="h-[240px]">
            <IncomeStreamsChart rows={s62} color={AMBER} retirementAge={62} />
          </div>
        </ChartCard>
        <ChartCard title={`Income Streams · Retire @ 67`} fixedHeight={false}>
          <div className="h-[240px]">
            <IncomeStreamsChart rows={s67} color={GREEN} retirementAge={67} />
          </div>
        </ChartCard>
      </div>

      {/* ── Break-Even Chart ── */}
      <ChartCard title="Cumulative Lifetime Income · Break-Even Comparison" className="mb-6" fixedHeight={false}>
        <div className="h-[280px]">
          <BreakEvenChart s62={s62} s67={s67} breakEvenAge={breakEvenAge} />
        </div>
      </ChartCard>

      {/* ── Projection Table ── */}
      <ProjectionTable s62={s62} s67={s67} currentAge={inputs.currentAge} />

      {/* ── Assumptions Footer ── */}
      <Card className="text-[11px] font-mono text-content-muted space-y-1">
        <p className="text-content-secondary font-semibold mb-2">Assumptions &amp; Methodology</p>
        <p>· All dollar figures shown in nominal (future) terms, not inflation-adjusted purchasing power.</p>
        <p>· 401K withdrawal uses the 4% rule: 4% of balance at retirement, inflated annually (COLA).</p>
        <p>· Social Security early claim (age 62) permanently reduced by 30% from FRA benefit.</p>
        <p>· SS benefits receive annual COLA equal to the inflation rate setting.</p>
        <p>· 401K contributions assumed to grow with income (at the inflation rate) each working year.</p>
        <p>· Break-even: first age where cumulative retirement income of age-67 scenario ≥ age-62 scenario.</p>
        <p>· This calculator is for illustrative purposes only. Consult a financial advisor for personalized guidance.</p>
      </Card>
    </div>
  )
}

export default Retirement
