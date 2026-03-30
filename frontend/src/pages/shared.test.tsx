import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import {
  SectionHeader,
  Card,
  ChartCard,
  Grid2,
  Grid3,
  Grid4,
  Th,
  Td,
  TickerCell,
  dirClass,
} from './shared'

describe('SectionHeader', () => {
  it('renders title', () => {
    render(<SectionHeader title="Market Overview" />)
    expect(screen.getByText('Market Overview')).toBeInTheDocument()
  })

  it('renders subtitle when provided', () => {
    render(<SectionHeader title="Overview" subtitle="Live data" />)
    expect(screen.getByText('Live data')).toBeInTheDocument()
  })

  it('does not render subtitle when absent', () => {
    const { container } = render(<SectionHeader title="Overview" />)
    expect(container.querySelector('span')).not.toBeInTheDocument()
  })
})

describe('Card', () => {
  it('renders children', () => {
    render(<Card>Hello</Card>)
    expect(screen.getByText('Hello')).toBeInTheDocument()
  })

  it('applies extra className', () => {
    const { container } = render(<Card className="custom-cls">X</Card>)
    expect(container.firstChild).toHaveClass('custom-cls')
  })
})

describe('ChartCard', () => {
  it('renders title and children', () => {
    render(<ChartCard title="My Chart"><span>chart</span></ChartCard>)
    expect(screen.getByText('My Chart')).toBeInTheDocument()
    expect(screen.getByText('chart')).toBeInTheDocument()
  })

  it('renders badge when provided', () => {
    render(<ChartCard title="Chart" badge={<span>badge-text</span>}><div /></ChartCard>)
    expect(screen.getByText('badge-text')).toBeInTheDocument()
  })
})

describe('Grid components', () => {
  it('Grid4 renders children', () => {
    render(<Grid4><div>child</div></Grid4>)
    expect(screen.getByText('child')).toBeInTheDocument()
  })

  it('Grid3 renders children', () => {
    render(<Grid3><div>child3</div></Grid3>)
    expect(screen.getByText('child3')).toBeInTheDocument()
  })

  it('Grid2 renders children', () => {
    render(<Grid2><div>child2</div></Grid2>)
    expect(screen.getByText('child2')).toBeInTheDocument()
  })
})

describe('Table components', () => {
  it('Table renders', () => {
    const { container } = render(<table><tbody><tr><Td>cell</Td></tr></tbody></table>)
    expect(container.querySelector('td')).toBeInTheDocument()
  })

  it('Th renders header text', () => {
    const { container } = render(<table><thead><tr><Th>Header</Th></tr></thead></table>)
    expect(container.querySelector('th')).toHaveTextContent('Header')
  })

  it('TickerCell renders ticker', () => {
    const { container } = render(<table><tbody><tr><TickerCell>AAPL</TickerCell></tr></tbody></table>)
    expect(container.querySelector('td')).toHaveTextContent('AAPL')
  })
})

describe('dirClass', () => {
  it('returns accent class for up', () => {
    expect(dirClass('up')).toBe('text-accent')
  })

  it('returns danger class for down', () => {
    expect(dirClass('down')).toBe('text-danger')
  })

  it('returns warn class for flat', () => {
    expect(dirClass('flat')).toBe('text-warn')
  })
})
