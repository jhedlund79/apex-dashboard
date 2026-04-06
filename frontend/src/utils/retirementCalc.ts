// ── Types ─────────────────────────────────────────────────────────────────────

export interface RetirementInputs {
  currentAge: number
  currentBalance: number
  annualIncome: number
  contributionRate: number
  returnRate: number
  inflationRate: number
  socialSecurityFRA: number
  lifeExpectancy: number
}

export interface YearData {
  age: number
  phase: 'accumulation' | 'retirement'
  balance: number
  ssIncome: number
  withdrawal: number
  contribution: number
  totalRetirementIncome: number
  cumulativeRetirementIncome: number
}

// ── Core Calculations ─────────────────────────────────────────────────────────

/**
 * Projects 401K balance and income streams from current age to life expectancy.
 *
 * Accumulation phase: contributions grow with inflation each year.
 * Retirement phase: 4% rule withdrawal and SS both COLA-adjusted annually.
 * SS at 62 is permanently reduced by 30% from the FRA benefit.
 */
export function computeScenario(inputs: RetirementInputs, retirementAge: 62 | 67): YearData[] {
  const {
    currentAge,
    currentBalance,
    annualIncome,
    contributionRate,
    returnRate,
    inflationRate,
    socialSecurityFRA,
    lifeExpectancy,
  } = inputs

  const ssBenefitBase = retirementAge === 62 ? socialSecurityFRA * 0.7 : socialSecurityFRA
  const baseContribution = annualIncome * contributionRate

  const rows: YearData[] = []
  let balance = currentBalance
  let cumulativeRetirementIncome = 0

  // Accumulation phase
  for (let age = currentAge; age < retirementAge; age++) {
    const yearsFromNow = age - currentAge
    const contribution = baseContribution * Math.pow(1 + inflationRate, yearsFromNow)
    balance = balance * (1 + returnRate) + contribution
    rows.push({
      age,
      phase: 'accumulation',
      balance,
      ssIncome: 0,
      withdrawal: 0,
      contribution,
      totalRetirementIncome: 0,
      cumulativeRetirementIncome: 0,
    })
  }

  // 4% rule: initial withdrawal anchored to balance at retirement
  const initialWithdrawal = balance * 0.04

  // Retirement phase
  for (let age = retirementAge; age <= lifeExpectancy; age++) {
    const yearsIn = age - retirementAge
    const ss = ssBenefitBase * Math.pow(1 + inflationRate, yearsIn)
    const desiredWithdrawal = initialWithdrawal * Math.pow(1 + inflationRate, yearsIn)

    // Apply investment return first, then deduct withdrawal — this ensures
    // the balance reaches exactly 0 when fully depleted rather than
    // asymptotically approaching zero.
    const grossBalance = balance * (1 + returnRate)
    const actualWithdrawal = grossBalance > 0 ? Math.min(desiredWithdrawal, grossBalance) : 0
    balance = Math.max(0, grossBalance - actualWithdrawal)
    const totalRetirementIncome = actualWithdrawal + ss
    cumulativeRetirementIncome += totalRetirementIncome

    rows.push({
      age,
      phase: 'retirement',
      balance,
      ssIncome: ss,
      withdrawal: actualWithdrawal,
      contribution: 0,
      totalRetirementIncome,
      cumulativeRetirementIncome,
    })
  }

  return rows
}

/**
 * Returns the first age at which the retire-67 scenario's cumulative retirement
 * income surpasses the retire-62 scenario's, or null if it never does.
 *
 * The retire-62 scenario has a 5-year head start; the retire-67 scenario has
 * higher annual income (larger 401K + full SS). The break-even is where the
 * higher annual payments finally overcome that head start.
 */
export function findBreakEvenAge(
  s62: YearData[],
  s67: YearData[],
  lifeExpectancy: number,
): number | null {
  const map62 = new Map(s62.map(d => [d.age, d]))
  const map67 = new Map(s67.map(d => [d.age, d]))

  for (let age = 67; age <= lifeExpectancy; age++) {
    const d62 = map62.get(age)
    const d67 = map67.get(age)
    if (d62 && d67 && d67.cumulativeRetirementIncome >= d62.cumulativeRetirementIncome) {
      return age
    }
  }
  return null
}

/**
 * Returns the age at which the 401K balance first hits zero, or null if it
 * is sustained through life expectancy.
 */
export function findDepletionAge(rows: YearData[]): number | null {
  for (const row of rows) {
    if (row.phase === 'retirement' && row.balance === 0) return row.age
  }
  return null
}

/** SS annual benefit for a given retirement age. */
export function ssAnnualBenefit(fra: number, retirementAge: 62 | 67): number {
  return retirementAge === 62 ? fra * 0.7 : fra
}
