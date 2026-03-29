import type { ReactNode, CSSProperties } from 'react'

interface SectionHeaderProps {
  title: string
  subtitle?: string
}

export const SectionHeader = ({ title, subtitle }: SectionHeaderProps) => (
  <div className="flex items-center justify-between mb-4.5">
    <h2 className="text-lg font-semibold flex items-center gap-2.5">
      {title}
      {subtitle && <span className="text-xs text-content-muted font-normal">{subtitle}</span>}
    </h2>
  </div>
)

interface GridProps {
  children: ReactNode
  className?: string
}

export const Grid4 = ({ children, className = 'mb-6' }: GridProps) => (
  <div className={`grid grid-cols-4 gap-4 ${className} max-[1200px]:grid-cols-2 max-[700px]:grid-cols-1`}>
    {children}
  </div>
)

export const Grid3 = ({ children, className = 'mb-6' }: GridProps) => (
  <div className={`grid grid-cols-3 gap-4 ${className} max-[1200px]:grid-cols-2 max-[700px]:grid-cols-1`}>
    {children}
  </div>
)

export const Grid2 = ({ children, className = 'mb-6' }: GridProps) => (
  <div className={`grid grid-cols-2 gap-4 ${className} max-[700px]:grid-cols-1`}>
    {children}
  </div>
)

interface CardProps {
  children: ReactNode
  className?: string
  style?: CSSProperties
}

export const Card = ({ children, className = '', style }: CardProps) => (
  <div
    style={style}
    className={`bg-surface border border-white/[0.04] rounded-card p-5 transition-all duration-300 animate-fadeIn hover:border-white/[0.08] hover:-translate-y-px ${className}`}
  >
    {children}
  </div>
)

interface ChartCardProps {
  title: string
  badge?: ReactNode
  children: ReactNode
  className?: string
  fixedHeight?: boolean
}

export const ChartCard = ({ title, badge, children, className = 'mb-6', fixedHeight = true }: ChartCardProps) => (
  <div className={`bg-surface border border-white/[0.04] rounded-card p-5 ${className}`}>
    <div className="flex items-center justify-between mb-3.5">
      <p className="font-mono text-[12px] uppercase tracking-[1.5px] text-content-muted">{title}</p>
      {badge}
    </div>
    <div className={fixedHeight ? 'relative h-[280px]' : ''}>{children}</div>
  </div>
)

interface TableProps {
  children: ReactNode
}

export const Table = ({ children }: TableProps) => (
  <table className="w-full border-collapse">{children}</table>
)

export const Th = ({ children }: { children: ReactNode }) => (
  <th className="font-mono text-[10px] uppercase tracking-[1.5px] text-content-muted text-left px-3.5 py-2.5 border-b border-white/[0.06]">
    {children}
  </th>
)

interface TdProps {
  children: ReactNode
  className?: string
}

export const Td = ({ children, className = '' }: TdProps) => (
  <td className={`px-3.5 py-3 border-b border-white/[0.03] text-[13px] ${className}`}>{children}</td>
)

export const TickerCell = ({ children }: { children: ReactNode }) => (
  <td className="px-3.5 py-3 border-b border-white/[0.03] font-mono font-semibold text-accent text-[13px]">
    {children}
  </td>
)

export const dirClass = (dir: string) =>
  dir === 'up' ? 'text-accent' : dir === 'down' ? 'text-danger' : 'text-warn'
