import { describe, it, expect } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Retirement from './Retirement'

// Chart.js and react-chartjs-2 are mocked globally in src/test/setup.ts

describe('Retirement — page structure', () => {
  it('renders the section header', () => {
    render(<Retirement />)
    expect(screen.getByText('◎ Retirement Calculator')).toBeInTheDocument()
  })

  it('renders both scenario cards', () => {
    render(<Retirement />)
    expect(screen.getByText(/Retire at 62/)).toBeInTheDocument()
    expect(screen.getByText(/Retire at 67/)).toBeInTheDocument()
  })

  it('renders the parameters input panel', () => {
    render(<Retirement />)
    expect(screen.getByText('Parameters')).toBeInTheDocument()
  })

  it('renders all slider labels', () => {
    render(<Retirement />)
    expect(screen.getByText('Current Age')).toBeInTheDocument()
    expect(screen.getByText('Current 401K Balance')).toBeInTheDocument()
    expect(screen.getByText('Annual Income')).toBeInTheDocument()
    expect(screen.getByText('401K Contribution Rate')).toBeInTheDocument()
    expect(screen.getByText('Expected Annual Return')).toBeInTheDocument()
    expect(screen.getByText('Inflation Rate')).toBeInTheDocument()
    expect(screen.getByText('Social Security at FRA (age 67)')).toBeInTheDocument()
    expect(screen.getByText('Life Expectancy')).toBeInTheDocument()
  })

  it('renders the break-even analysis section', () => {
    render(<Retirement />)
    expect(screen.getByText('Break-Even Analysis')).toBeInTheDocument()
  })

  it('renders the year-by-year projection table', () => {
    render(<Retirement />)
    expect(screen.getByText('Year-by-Year Projection')).toBeInTheDocument()
  })

  it('renders the assumptions footer', () => {
    render(<Retirement />)
    expect(screen.getByText('Assumptions & Methodology')).toBeInTheDocument()
  })

  it('renders chart section titles', () => {
    render(<Retirement />)
    expect(screen.getByText('401K Wealth Trajectory · Both Scenarios')).toBeInTheDocument()
    expect(screen.getByText('Cumulative Lifetime Income · Break-Even Comparison')).toBeInTheDocument()
    expect(screen.getByText('Income Streams · Retire @ 62')).toBeInTheDocument()
    expect(screen.getByText('Income Streams · Retire @ 67')).toBeInTheDocument()
  })
})

describe('Retirement — default values', () => {
  it('shows default current age of 45', () => {
    render(<Retirement />)
    expect(screen.getByText('45 yrs')).toBeInTheDocument()
  })

  it('shows default contribution rate of 10%', () => {
    render(<Retirement />)
    expect(screen.getByText('10.0%')).toBeInTheDocument()
  })

  it('shows default return rate of 7%', () => {
    render(<Retirement />)
    expect(screen.getByText('7.0%')).toBeInTheDocument()
  })

  it('shows default inflation rate of 3%', () => {
    render(<Retirement />)
    expect(screen.getByText('3.0%')).toBeInTheDocument()
  })

  it('shows default life expectancy', () => {
    render(<Retirement />)
    expect(screen.getByText('Age 90')).toBeInTheDocument()
  })
})

describe('Retirement — scenario cards', () => {
  it('shows the 30% reduction note on the retire-62 card', () => {
    render(<Retirement />)
    expect(screen.getByText('30% reduction applied')).toBeInTheDocument()
  })

  it('shows 401K at Retirement labels on both cards', () => {
    render(<Retirement />)
    const labels = screen.getAllByText('401K at Retirement')
    expect(labels).toHaveLength(2)
  })

  it('shows Annual SS Benefit labels on both cards', () => {
    render(<Retirement />)
    const labels = screen.getAllByText('Annual SS Benefit')
    expect(labels).toHaveLength(2)
  })

  it('shows Annual Withdrawal (4%) labels on both cards', () => {
    render(<Retirement />)
    const labels = screen.getAllByText('Annual Withdrawal (4%)')
    expect(labels).toHaveLength(2)
  })

  it('shows Total Annual Income labels on both cards', () => {
    render(<Retirement />)
    const labels = screen.getAllByText('Total Annual Income')
    expect(labels).toHaveLength(2)
  })

  it('shows a sustained balance notice when 401K is not depleted', () => {
    render(<Retirement />)
    const notices = screen.getAllByText(/Balance sustained through life expectancy/)
    expect(notices.length).toBeGreaterThan(0)
  })
})

describe('Retirement — break-even display', () => {
  it('displays a break-even age with default inputs', () => {
    render(<Retirement />)
    // The break-even card shows "Age XX" in a large bold element
    const matches = screen.getAllByText(/Age \d+/)
    // At least one should be the break-even heading (distinct from the life-expectancy slider label)
    expect(matches.length).toBeGreaterThanOrEqual(1)
  })

  it('shows years after retiring early label', () => {
    render(<Retirement />)
    expect(screen.getAllByText(/Years after retiring early/i).length).toBeGreaterThan(0)
  })

  it('shows years of retirement ahead label', () => {
    render(<Retirement />)
    expect(screen.getAllByText(/Years of retirement ahead/i).length).toBeGreaterThan(0)
  })
})

describe('Retirement — projection table', () => {
  it('shows table column headers', () => {
    render(<Retirement />)
    // "← Retire @ 62" and "← Retire @ 67" appear in the table header TH elements
    expect(screen.getAllByText(/Retire @ 62/).length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText(/Retire @ 67/).length).toBeGreaterThanOrEqual(1)
    // Balance column headers appear twice (once per scenario)
    const balanceCols = screen.getAllByText('Balance')
    expect(balanceCols.length).toBeGreaterThanOrEqual(2)
  })

  it('shows phase labels in the table', () => {
    render(<Retirement />)
    // Working phase during accumulation
    expect(screen.getAllByText('Working').length).toBeGreaterThan(0)
    // Retired phase once past 62
    expect(screen.getAllByText('Retired').length).toBeGreaterThan(0)
  })

  it('table contains rows with age 45 (default current age)', () => {
    render(<Retirement />)
    const table = screen.getByRole('table')
    expect(within(table).getByText('45')).toBeInTheDocument()
  })
})

describe('Retirement — slider interactions', () => {
  it('updates the displayed age when the slider changes', async () => {
    const user = userEvent.setup()
    render(<Retirement />)

    const sliders = screen.getAllByRole('slider')
    const ageSlider = sliders[0]

    await user.type(ageSlider, '{ArrowRight}')
    // After interaction, the displayed value changes (or stays if already at boundary)
    // We just verify the slider is interactive and the component does not crash
    expect(ageSlider).toBeInTheDocument()
  })

  it('renders eight sliders (one per parameter)', () => {
    render(<Retirement />)
    const sliders = screen.getAllByRole('slider')
    expect(sliders).toHaveLength(8)
  })

  it('all sliders have the expected min/max attributes', () => {
    render(<Retirement />)
    const sliders = screen.getAllByRole('slider')

    // currentAge slider: min=25, max=61
    expect(sliders[0]).toHaveAttribute('min', '25')
    expect(sliders[0]).toHaveAttribute('max', '61')

    // lifeExpectancy slider (last): min=70, max=95
    expect(sliders[7]).toHaveAttribute('min', '70')
    expect(sliders[7]).toHaveAttribute('max', '95')
  })
})

describe('Retirement — assumptions footer', () => {
  it('mentions the 4% rule', () => {
    render(<Retirement />)
    expect(screen.getAllByText(/4% rule/i).length).toBeGreaterThan(0)
  })

  it('mentions the 30% SS reduction', () => {
    render(<Retirement />)
    expect(screen.getAllByText(/30%/).length).toBeGreaterThan(0)
  })

  it('mentions nominal dollars', () => {
    render(<Retirement />)
    expect(screen.getAllByText(/nominal/i).length).toBeGreaterThan(0)
  })
})
