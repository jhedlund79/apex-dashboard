interface HeaderProps {
  fearGreed?: string
  marketStatus?: string
}

export const Header = ({ fearGreed, marketStatus }: HeaderProps) => (
  <header className="col-span-full px-8 py-5 flex items-center justify-between border-b border-white/[0.04] bg-gradient-to-b from-accent/[0.02] to-transparent">
    <div className="flex items-center gap-3">
      <div className="w-9 h-9 bg-gradient-to-br from-accent to-brand-cyan rounded-lg flex items-center justify-center font-bold text-bg text-base">
        A
      </div>
      <div>
        <div className="font-mono text-lg font-medium tracking-[3px] text-accent">APEX</div>
        <div className="text-[11px] text-content-muted tracking-wider mt-0.5">
          Aggressive Growth Terminal
        </div>
      </div>
    </div>

    <div className="flex items-center gap-6">
      {fearGreed && (
        <div className="px-3.5 py-1.5 rounded font-mono text-xs bg-danger/[0.12] text-danger border border-danger/20">
          {fearGreed}
        </div>
      )}
      <div className="flex items-center gap-2 font-mono text-xs text-content-secondary">
        <span className="w-2 h-2 rounded-full bg-accent animate-pulse" />
        {marketStatus ?? '—'}
      </div>
    </div>
  </header>
)
