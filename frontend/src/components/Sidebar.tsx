import type { QuickStat } from '../types/market'

const navItems = [
  { id: 'overview', icon: '◉', label: 'Overview', section: 'Markets' },
  { id: 'us', icon: '▣', label: 'US Markets', section: 'Markets' },
  { id: 'intl', icon: '◈', label: 'International', section: 'Markets' },
  { id: 'crypto', icon: '⬡', label: 'Crypto', section: 'Markets' },
  { id: 'recs', icon: '⚡', label: 'Recommendations', section: 'Strategy' },
  { id: 'sectors', icon: '◧', label: 'Sector Analysis', section: 'Strategy' },
  { id: 'principal', icon: '◎', label: 'Principal 401(k)', section: 'Portfolio' },
  { id: 'morganstanley', icon: '◈', label: 'Morgan Stanley', section: 'Portfolio' },
  { id: 'fidelity', icon: '◆', label: 'Fidelity', section: 'Portfolio' },
  { id: 'sofi', icon: '◇', label: 'SoFi Invest', section: 'Portfolio' },
] as const

export type TabId = (typeof navItems)[number]['id']

interface SidebarProps {
  activeTab: TabId
  onTabChange: (id: TabId) => void
  quickStats?: QuickStat[]
}

const statColor = (dir: QuickStat['changeDir']) =>
  dir === 'up' ? 'text-warn' : dir === 'down' ? 'text-danger' : 'text-content-primary'

export const Sidebar = ({ activeTab, onTabChange, quickStats }: SidebarProps) => {
  const markets = navItems.filter(n => n.section === 'Markets')
  const strategy = navItems.filter(n => n.section === 'Strategy')
  const portfolio = navItems.filter(n => n.section === 'Portfolio')

  return (
    <nav className="bg-surface border-r border-white/[0.04] py-6 overflow-y-auto">
      <NavSection title="Markets">
        {markets.map(item => (
          <NavItem key={item.id} item={item} active={activeTab === item.id} onClick={() => onTabChange(item.id)} />
        ))}
      </NavSection>

      <NavSection title="Strategy">
        {strategy.map(item => (
          <NavItem key={item.id} item={item} active={activeTab === item.id} onClick={() => onTabChange(item.id)} />
        ))}
      </NavSection>

      <NavSection title="Portfolio">
        {portfolio.map(item => (
          <NavItem key={item.id} item={item} active={activeTab === item.id} onClick={() => onTabChange(item.id)} />
        ))}
      </NavSection>

      <NavSection title="Quick Stats">
        <div className="px-3 py-2">
          {quickStats?.map(stat => (
            <div key={stat.label} className="mb-4">
              <div className="text-[11px] text-content-muted font-mono mb-1">{stat.label}</div>
              <div className={`text-[22px] font-bold ${statColor(stat.changeDir)}`}>{stat.value}</div>
              <div className={`text-[11px] font-mono ${statColor(stat.changeDir)}`}>{stat.change}</div>
            </div>
          ))}
        </div>
      </NavSection>
    </nav>
  )
}

const NavSection = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div className="px-5 mb-7">
    <div className="font-mono text-[10px] uppercase tracking-[2px] text-content-muted mb-3.5 pb-2 border-b border-white/[0.04]">
      {title}
    </div>
    {children}
  </div>
)

const NavItem = ({
  item,
  active,
  onClick,
}: {
  item: { icon: string; label: string }
  active: boolean
  onClick: () => void
}) => (
  <div
    onClick={onClick}
    className={`flex items-center gap-2.5 px-3 py-2.5 rounded-lg cursor-pointer transition-all duration-200 mb-0.5 text-[13px] ${
      active
        ? 'bg-accent/[0.08] text-accent'
        : 'text-content-secondary hover:bg-white/[0.04] hover:text-content-primary'
    }`}
  >
    <span className="w-4.5 text-center text-sm">{item.icon}</span>
    {item.label}
  </div>
)
