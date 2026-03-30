import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Sidebar } from './Sidebar'
import type { TabId } from './Sidebar'

const noop = vi.fn()

describe('Sidebar', () => {
  it('renders all nav sections', () => {
    render(<Sidebar activeTab="overview" onTabChange={noop} />)
    expect(screen.getByText('Markets')).toBeInTheDocument()
    expect(screen.getByText('Strategy')).toBeInTheDocument()
    expect(screen.getByText('Portfolio')).toBeInTheDocument()
  })

  it('renders all nav items', () => {
    render(<Sidebar activeTab="overview" onTabChange={noop} />)
    expect(screen.getByText('Overview')).toBeInTheDocument()
    expect(screen.getByText('US Markets')).toBeInTheDocument()
    expect(screen.getByText('International')).toBeInTheDocument()
    expect(screen.getByText('Crypto')).toBeInTheDocument()
    expect(screen.getByText('Recommendations')).toBeInTheDocument()
    expect(screen.getByText('Sector Analysis')).toBeInTheDocument()
    expect(screen.getByText('Principal 401(k)')).toBeInTheDocument()
    expect(screen.getByText('Morgan Stanley')).toBeInTheDocument()
    expect(screen.getByText('Fidelity')).toBeInTheDocument()
    expect(screen.getByText('SoFi Invest')).toBeInTheDocument()
  })

  it('calls onTabChange when nav item is clicked', () => {
    const onChange = vi.fn()
    render(<Sidebar activeTab="overview" onTabChange={onChange} />)
    fireEvent.click(screen.getByText('Crypto'))
    expect(onChange).toHaveBeenCalledWith('crypto')
  })

  it('shows quick stats when provided', () => {
    const quickStats = [
      { label: 'BTC', value: '$82k', change: '+2%', changeDir: 'up' as const },
      { label: 'ETH', value: '$3k', change: '-1%', changeDir: 'down' as const },
    ]
    render(<Sidebar activeTab="overview" onTabChange={noop} quickStats={quickStats} />)
    expect(screen.getByText('BTC')).toBeInTheDocument()
    expect(screen.getByText('$82k')).toBeInTheDocument()
    expect(screen.getByText('ETH')).toBeInTheDocument()
  })

  it('renders Quick Stats section header', () => {
    render(<Sidebar activeTab="overview" onTabChange={noop} />)
    expect(screen.getByText('Quick Stats')).toBeInTheDocument()
  })

  it.each<[TabId, string]>([
    ['overview', 'Overview'],
    ['us', 'US Markets'],
    ['crypto', 'Crypto'],
    ['principal', 'Principal 401(k)'],
    ['morganstanley', 'Morgan Stanley'],
    ['fidelity', 'Fidelity'],
    ['sofi', 'SoFi Invest'],
  ])('onTabChange called with %s when %s is clicked', (tabId, label) => {
    const onChange = vi.fn()
    render(<Sidebar activeTab="overview" onTabChange={onChange} />)
    fireEvent.click(screen.getByText(label))
    expect(onChange).toHaveBeenCalledWith(tabId)
  })
})
