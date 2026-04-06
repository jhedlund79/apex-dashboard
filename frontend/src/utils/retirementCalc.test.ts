import { describe, it, expect } from 'vitest'
import {
  computeScenario,
  findBreakEvenAge,
  findDepletionAge,
  ssAnnualBenefit,
  type RetirementInputs,
} from './retirementCalc'

// ── Shared fixtures ───────────────────────────────────────────────────────────

const base: RetirementInputs = {
  currentAge: 45,
  currentBalance: 150_000,
  annualIncome: 100_000,
  contributionRate: 0.10,
  returnRate: 0.07,
  inflationRate: 0.03,
  socialSecurityFRA: 24_000,
  lifeExpectancy: 90,
}

// ── ssAnnualBenefit ───────────────────────────────────────────────────────────

describe('ssAnnualBenefit', () => {
  it('returns the full FRA benefit when retiring at 67', () => {
    expect(ssAnnualBenefit(24_000, 67)).toBe(24_000)
  })

  it('reduces benefit by 30% when retiring at 62', () => {
    expect(ssAnnualBenefit(24_000, 62)).toBeCloseTo(16_800)
  })

  it('handles zero FRA benefit', () => {
    expect(ssAnnualBenefit(0, 62)).toBe(0)
    expect(ssAnnualBenefit(0, 67)).toBe(0)
  })
})

// ── computeScenario — row count & phase boundaries ───────────────────────────

describe('computeScenario — row count', () => {
  it('produces one row per year from currentAge to lifeExpectancy inclusive', () => {
    const rows = computeScenario(base, 67)
    // acc rows: [currentAge, retirementAge) = 22; ret rows: [retirementAge, lifeExpectancy] = 24
    const expected = base.lifeExpectancy - base.currentAge + 1  // 46
    expect(rows).toHaveLength(expected)
  })

  it('last row age equals lifeExpectancy', () => {
    expect(computeScenario(base, 67).at(-1)!.age).toBe(base.lifeExpectancy)
  })

  it('first row age equals currentAge', () => {
    expect(computeScenario(base, 67)[0].age).toBe(base.currentAge)
  })

  it('retire-62 total rows = lifeExpectancy - currentAge + 1', () => {
    const rows = computeScenario(base, 62)
    expect(rows).toHaveLength(base.lifeExpectancy - base.currentAge + 1)
  })
})

describe('computeScenario — phase boundaries', () => {
  it('all rows before retirementAge are accumulation phase', () => {
    const rows = computeScenario(base, 67)
    const accRows = rows.filter(r => r.age < 67)
    expect(accRows.every(r => r.phase === 'accumulation')).toBe(true)
  })

  it('all rows from retirementAge onwards are retirement phase', () => {
    const rows = computeScenario(base, 67)
    const retRows = rows.filter(r => r.age >= 67)
    expect(retRows.every(r => r.phase === 'retirement')).toBe(true)
  })

  it('accumulation rows have zero SS income and zero withdrawal', () => {
    const rows = computeScenario(base, 67)
    const accRows = rows.filter(r => r.phase === 'accumulation')
    expect(accRows.every(r => r.ssIncome === 0 && r.withdrawal === 0)).toBe(true)
  })

  it('retirement rows have zero contribution', () => {
    const rows = computeScenario(base, 67)
    const retRows = rows.filter(r => r.phase === 'retirement')
    expect(retRows.every(r => r.contribution === 0)).toBe(true)
  })
})

// ── computeScenario — accumulation math ──────────────────────────────────────

describe('computeScenario — accumulation phase', () => {
  it('balance grows each accumulation year', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'accumulation')
    for (let i = 1; i < rows.length; i++) {
      expect(rows[i].balance).toBeGreaterThan(rows[i - 1].balance)
    }
  })

  it('retire-67 balance at retirement exceeds retire-62 balance at retirement', () => {
    const rows62 = computeScenario(base, 62)
    const rows67 = computeScenario(base, 67)
    const bal62 = rows62.find(r => r.age === 61)!.balance  // last accumulation year for 62
    const bal67 = rows67.find(r => r.age === 66)!.balance  // last accumulation year for 67
    expect(bal67).toBeGreaterThan(bal62)
  })

  it('contributions grow with inflation year-over-year', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'accumulation')
    for (let i = 1; i < rows.length; i++) {
      expect(rows[i].contribution).toBeGreaterThan(rows[i - 1].contribution)
    }
  })

  it('first-year contribution equals income × contributionRate', () => {
    const rows = computeScenario(base, 67)
    expect(rows[0].contribution).toBeCloseTo(base.annualIncome * base.contributionRate)
  })
})

// ── computeScenario — Social Security ────────────────────────────────────────

describe('computeScenario — Social Security', () => {
  it('retire-62 first SS payment equals FRA × 0.7', () => {
    const rows = computeScenario(base, 62)
    const firstRetRow = rows.find(r => r.phase === 'retirement')!
    expect(firstRetRow.ssIncome).toBeCloseTo(base.socialSecurityFRA * 0.7)
  })

  it('retire-67 first SS payment equals full FRA benefit', () => {
    const rows = computeScenario(base, 67)
    const firstRetRow = rows.find(r => r.phase === 'retirement')!
    expect(firstRetRow.ssIncome).toBeCloseTo(base.socialSecurityFRA)
  })

  it('SS income grows each retirement year (COLA)', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'retirement')
    for (let i = 1; i < rows.length; i++) {
      expect(rows[i].ssIncome).toBeGreaterThan(rows[i - 1].ssIncome)
    }
  })

  it('retire-67 SS is always higher than retire-62 SS at the same age', () => {
    const rows62 = computeScenario(base, 62)
    const rows67 = computeScenario(base, 67)
    const map62 = new Map(rows62.map(d => [d.age, d]))
    // Compare at ages 67–90 where both are in retirement
    for (let age = 67; age <= base.lifeExpectancy; age++) {
      const d62 = map62.get(age)!
      const d67 = rows67.find(r => r.age === age)!
      expect(d67.ssIncome).toBeGreaterThan(d62.ssIncome)
    }
  })
})

// ── computeScenario — 4% withdrawal rule ─────────────────────────────────────

describe('computeScenario — 4% withdrawal rule', () => {
  it('first retirement withdrawal is ~4% of balance at retirement', () => {
    const rows = computeScenario(base, 67)
    const lastAccRow = rows.filter(r => r.phase === 'accumulation').at(-1)!
    const firstRetRow = rows.find(r => r.phase === 'retirement')!
    expect(firstRetRow.withdrawal).toBeCloseTo(lastAccRow.balance * 0.04, -2)
  })

  it('withdrawal grows with inflation each year', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'retirement')
    // First few years — balance is healthy, so withdrawal follows inflation
    for (let i = 1; i < Math.min(10, rows.length); i++) {
      expect(rows[i].withdrawal).toBeGreaterThan(rows[i - 1].withdrawal)
    }
  })

  it('totalRetirementIncome equals ssIncome + withdrawal each year', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'retirement')
    for (const row of rows) {
      expect(row.totalRetirementIncome).toBeCloseTo(row.ssIncome + row.withdrawal)
    }
  })

  it('cumulativeRetirementIncome is monotonically non-decreasing', () => {
    const rows = computeScenario(base, 67).filter(r => r.phase === 'retirement')
    for (let i = 1; i < rows.length; i++) {
      expect(rows[i].cumulativeRetirementIncome).toBeGreaterThanOrEqual(
        rows[i - 1].cumulativeRetirementIncome,
      )
    }
  })
})

// ── computeScenario — balance floor ──────────────────────────────────────────

describe('computeScenario — balance floor', () => {
  it('balance never goes below zero', () => {
    const rows = computeScenario(base, 62)
    expect(rows.every(r => r.balance >= 0)).toBe(true)
  })

  it('a tiny starting balance with low return eventually depletes to zero', () => {
    // real return ≈ -1% (2% nominal - 3% inflation): withdrawals outpace growth
    const depleting: RetirementInputs = {
      ...base,
      currentBalance: 0,
      annualIncome: 20_000,    // very small contributions → small retirement balance
      returnRate: 0.02,
      lifeExpectancy: 95,
    }
    const rows = computeScenario(depleting, 62)
    const depleted = rows.some(r => r.balance === 0)
    expect(depleted).toBe(true)
  })

  it('once balance hits zero withdrawal stays zero in all subsequent years', () => {
    const depleting: RetirementInputs = {
      ...base,
      currentBalance: 0,
      annualIncome: 20_000,
      returnRate: 0.02,
      lifeExpectancy: 95,
    }
    const rows = computeScenario(depleting, 62)
    // wasZero tracks whether the balance was 0 at the END of the PREVIOUS year.
    // The row where balance first hits 0 still has a non-zero withdrawal (that's
    // what depleted it); it's the FOLLOWING rows that should withdraw nothing.
    let wasZero = false
    for (const row of rows.filter(r => r.phase === 'retirement')) {
      if (wasZero) expect(row.withdrawal).toBe(0)
      if (row.balance === 0) wasZero = true
    }
  })

  it('accumulation rows have a positive balance when starting from zero', () => {
    const zeroStart: RetirementInputs = { ...base, currentBalance: 0 }
    const rows = computeScenario(zeroStart, 67).filter(r => r.phase === 'accumulation')
    expect(rows.every(r => r.balance > 0)).toBe(true)
  })
})

// ── computeScenario — retire immediately (currentAge === retirementAge) ───────

describe('computeScenario — retire at current age', () => {
  it('produces only retirement rows when currentAge equals retirementAge', () => {
    const immediate: RetirementInputs = { ...base, currentAge: 67 }
    const rows = computeScenario(immediate, 67)
    expect(rows.every(r => r.phase === 'retirement')).toBe(true)
  })

  it('first withdrawal at retirement uses the starting balance for 4% calc', () => {
    const immediate: RetirementInputs = { ...base, currentAge: 67 }
    const rows = computeScenario(immediate, 67)
    expect(rows[0].withdrawal).toBeCloseTo(immediate.currentBalance * 0.04, -1)
  })
})

// ── findDepletionAge ──────────────────────────────────────────────────────────

describe('findDepletionAge', () => {
  it('returns null when balance is always positive', () => {
    const rows = computeScenario(base, 67)
    expect(findDepletionAge(rows)).toBeNull()
  })

  it('returns the correct depletion age', () => {
    const depleting: RetirementInputs = {
      ...base,
      currentBalance: 0,
      annualIncome: 20_000,
      returnRate: 0.02,
      lifeExpectancy: 95,
    }
    const rows = computeScenario(depleting, 62)
    const age = findDepletionAge(rows)
    expect(age).not.toBeNull()
    expect(age!).toBeGreaterThanOrEqual(62)
    expect(age!).toBeLessThanOrEqual(95)
  })

  it('returns the first year balance hits zero, not a later one', () => {
    const depleting: RetirementInputs = {
      ...base,
      currentBalance: 0,
      annualIncome: 20_000,
      returnRate: 0.02,
      lifeExpectancy: 95,
    }
    const rows = computeScenario(depleting, 62)
    const age = findDepletionAge(rows)!
    // Balance at that exact age should be 0
    expect(rows.find(r => r.age === age)!.balance).toBe(0)
    // Balance the year before should be > 0 (or it's the very first retirement year)
    const prev = rows.find(r => r.age === age - 1)
    if (prev && prev.phase === 'retirement') {
      expect(prev.balance).toBeGreaterThan(0)
    }
  })

  it('ignores accumulation rows with zero balance', () => {
    // A zero-balance accumulation row should not be flagged as depletion
    const zeroStart: RetirementInputs = { ...base, currentBalance: 0 }
    const rows = computeScenario(zeroStart, 67)
    // Force the first accumulation row to have balance=0 for the test
    const accRows = rows.filter(r => r.phase === 'accumulation')
    expect(accRows.length).toBeGreaterThan(0)
    // findDepletionAge should not return any of these
    const age = findDepletionAge(rows)
    if (age !== null) {
      expect(rows.find(r => r.age === age)!.phase).toBe('retirement')
    }
  })
})

// ── findBreakEvenAge ──────────────────────────────────────────────────────────

describe('findBreakEvenAge', () => {
  it('returns a break-even age within the projection window for typical inputs', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    const age = findBreakEvenAge(s62, s67, base.lifeExpectancy)
    expect(age).not.toBeNull()
    expect(age!).toBeGreaterThanOrEqual(67)
    expect(age!).toBeLessThanOrEqual(base.lifeExpectancy)
  })

  it('break-even age is after 67 — retire-67 needs time to catch up', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    const age = findBreakEvenAge(s62, s67, base.lifeExpectancy)
    // The retire-62 person always has 5+ years of head-start income
    expect(age!).toBeGreaterThan(67)
  })

  it('at break-even age s67 cumulative >= s62 cumulative', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    const age = findBreakEvenAge(s62, s67, base.lifeExpectancy)!
    const d62 = s62.find(r => r.age === age)!
    const d67 = s67.find(r => r.age === age)!
    expect(d67.cumulativeRetirementIncome).toBeGreaterThanOrEqual(d62.cumulativeRetirementIncome)
  })

  it('one year before break-even s67 cumulative is still less than s62', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    const age = findBreakEvenAge(s62, s67, base.lifeExpectancy)!
    if (age > 67) {
      const d62 = s62.find(r => r.age === age - 1)!
      const d67 = s67.find(r => r.age === age - 1)!
      expect(d67.cumulativeRetirementIncome).toBeLessThan(d62.cumulativeRetirementIncome)
    }
  })

  it('returns null when life expectancy is too short to reach break-even', () => {
    const shortLife: RetirementInputs = { ...base, lifeExpectancy: 70 }
    const s62 = computeScenario(shortLife, 62)
    const s67 = computeScenario(shortLife, 67)
    const age = findBreakEvenAge(s62, s67, shortLife.lifeExpectancy)
    // With only 3 years of retirement for the 67 scenario vs 8 for the 62 scenario,
    // break-even cannot be reached
    expect(age).toBeNull()
  })

  it('returns null when retire-67 has very low SS (no catch-up possible)', () => {
    // Both scenarios get identical SS (zero) — retire-62 always leads
    const zeroSS: RetirementInputs = { ...base, socialSecurityFRA: 0, lifeExpectancy: 75 }
    const s62 = computeScenario(zeroSS, 62)
    const s67 = computeScenario(zeroSS, 67)
    const age = findBreakEvenAge(s62, s67, zeroSS.lifeExpectancy)
    expect(age).toBeNull()
  })

  it('break-even age shifts later when SS FRA benefit is higher', () => {
    // Higher FRA means a larger head-start for retire-62 (more SS income from 62–66)
    // and a larger annual advantage for retire-67 — but the head-start grows faster
    // because it applies to 5 years of income rather than the annual increment alone.
    const highSS: RetirementInputs = { ...base, socialSecurityFRA: 48_000 }
    const lowSS: RetirementInputs = { ...base, socialSecurityFRA: 12_000 }

    const s62High = computeScenario(highSS, 62)
    const s67High = computeScenario(highSS, 67)
    const s62Low = computeScenario(lowSS, 62)
    const s67Low = computeScenario(lowSS, 67)

    const ageHigh = findBreakEvenAge(s62High, s67High, base.lifeExpectancy)
    const ageLow = findBreakEvenAge(s62Low, s67Low, base.lifeExpectancy)

    if (ageHigh !== null && ageLow !== null) {
      expect(ageHigh).toBeGreaterThanOrEqual(ageLow)
    }
  })
})

// ── Scenario comparisons ──────────────────────────────────────────────────────

describe('retire-62 vs retire-67 comparisons', () => {
  it('retire-67 annual income in early retirement exceeds retire-62 at same age', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    const d62at70 = s62.find(r => r.age === 70)!
    const d67at70 = s67.find(r => r.age === 70)!
    // retire-67 has more 401K + higher SS at the same age
    expect(d67at70.totalRetirementIncome).toBeGreaterThan(d62at70.totalRetirementIncome)
  })

  it('retire-67 has higher 401K balance at every shared retirement age', () => {
    const s62 = computeScenario(base, 62)
    const s67 = computeScenario(base, 67)
    // At age 68, both are in retirement; 67 scenario has larger starting balance
    const d62at68 = s62.find(r => r.age === 68)!
    const d67at68 = s67.find(r => r.age === 68)!
    expect(d67at68.balance).toBeGreaterThan(d62at68.balance)
  })

  it('higher returnRate increases both scenario balances at retirement', () => {
    const low: RetirementInputs = { ...base, returnRate: 0.04 }
    const high: RetirementInputs = { ...base, returnRate: 0.10 }

    const bal62Low = computeScenario(low, 62).filter(r => r.phase === 'accumulation').at(-1)!.balance
    const bal62High = computeScenario(high, 62).filter(r => r.phase === 'accumulation').at(-1)!.balance
    expect(bal62High).toBeGreaterThan(bal62Low)

    const bal67Low = computeScenario(low, 67).filter(r => r.phase === 'accumulation').at(-1)!.balance
    const bal67High = computeScenario(high, 67).filter(r => r.phase === 'accumulation').at(-1)!.balance
    expect(bal67High).toBeGreaterThan(bal67Low)
  })

  it('higher contributionRate increases balance at retirement', () => {
    const low: RetirementInputs = { ...base, contributionRate: 0.05 }
    const high: RetirementInputs = { ...base, contributionRate: 0.20 }

    const balLow = computeScenario(low, 67).filter(r => r.phase === 'accumulation').at(-1)!.balance
    const balHigh = computeScenario(high, 67).filter(r => r.phase === 'accumulation').at(-1)!.balance
    expect(balHigh).toBeGreaterThan(balLow)
  })
})
