import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { LoadingState } from './LoadingState'

describe('LoadingState', () => {
  it('renders 6 skeleton cards', () => {
    const { container } = render(<LoadingState />)
    const cards = container.querySelectorAll('.animate-shimmer')
    expect(cards).toHaveLength(6)
  })

  it('applies animation delays to each card', () => {
    const { container } = render(<LoadingState />)
    const cards = container.querySelectorAll('.animate-shimmer')
    expect((cards[0] as HTMLElement).style.animationDelay).toBe('0s')
    expect((cards[1] as HTMLElement).style.animationDelay).toBe('0.1s')
    expect((cards[5] as HTMLElement).style.animationDelay).toBe('0.5s')
  })
})
