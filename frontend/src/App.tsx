import { lazy, Suspense, useState } from 'react'
import { Header } from './components/Header'
import { Sidebar, type TabId } from './components/Sidebar'
import { LoadingState } from './components/LoadingState'
import { useOverview } from './hooks/useMarketData'

const Overview = lazy(() => import('./pages/Overview'))
const USMarkets = lazy(() => import('./pages/USMarkets'))
const International = lazy(() => import('./pages/International'))
const Crypto = lazy(() => import('./pages/Crypto'))
const Recommendations = lazy(() => import('./pages/Recommendations'))
const Sectors = lazy(() => import('./pages/Sectors'))
const Principal = lazy(() => import('./pages/Principal'))
const MorganStanley = lazy(() => import('./pages/MorganStanley'))
const Fidelity = lazy(() => import('./pages/Fidelity'))
const SoFi = lazy(() => import('./pages/SoFi'))
const Simulation = lazy(() => import('./pages/Simulation'))

const TABS: Record<TabId, React.LazyExoticComponent<() => React.JSX.Element>> = {
  overview: Overview,
  us: USMarkets,
  intl: International,
  crypto: Crypto,
  recs: Recommendations,
  sectors: Sectors,
  principal: Principal,
  morganstanley: MorganStanley,
  fidelity: Fidelity,
  sofi: SoFi,
  simulation: Simulation,
}

const App = () => {
  const [activeTab, setActiveTab] = useState<TabId>('overview')
  const { data: overviewData } = useOverview()
  const TabComponent = TABS[activeTab]

  return (
    <div className="grid grid-cols-[280px_1fr] grid-rows-[auto_1fr] min-h-screen bg-bg text-content-primary font-sans">
      <Header
        fearGreed={overviewData?.header?.fearGreed}
        marketStatus={overviewData?.header?.marketStatus}
      />
      <Sidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        quickStats={overviewData?.quickStats}
      />
      <main className="p-6 overflow-y-auto max-h-[calc(100vh-72px)]">
        <Suspense fallback={<LoadingState />}>
          <TabComponent />
        </Suspense>
      </main>
    </div>
  )
}

export default App
