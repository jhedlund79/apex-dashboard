import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Header } from './Header'

describe('Header', () => {
  it('renders APEX brand text', () => {
    render(<Header />)
    expect(screen.getByText('APEX')).toBeInTheDocument()
    expect(screen.getByText('Aggressive Growth Terminal')).toBeInTheDocument()
  })

  it('shows market status when provided', () => {
    render(<Header marketStatus="Open" />)
    expect(screen.getByText('Open')).toBeInTheDocument()
  })

  it('shows dash when market status is absent', () => {
    render(<Header />)
    expect(screen.getByText('—')).toBeInTheDocument()
  })

  it('shows fear/greed indicator when provided', () => {
    render(<Header fearGreed="Extreme Fear" />)
    expect(screen.getByText('Extreme Fear')).toBeInTheDocument()
  })

  it('does not render fear/greed element when absent', () => {
    render(<Header />)
    expect(screen.queryByText('Extreme Fear')).not.toBeInTheDocument()
  })

  it('renders A logo mark', () => {
    render(<Header />)
    expect(screen.getByText('A')).toBeInTheDocument()
  })
})
