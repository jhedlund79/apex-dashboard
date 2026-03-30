import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorState } from './ErrorState'

describe('ErrorState', () => {
  it('renders error heading', () => {
    render(<ErrorState error={null} />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('shows error message when provided', () => {
    render(<ErrorState error={new Error('Connection refused')} />)
    expect(screen.getByText('Connection refused')).toBeInTheDocument()
  })

  it('shows Unknown error when error is null', () => {
    render(<ErrorState error={null} />)
    expect(screen.getByText('Unknown error')).toBeInTheDocument()
  })

  it('renders retry button when onRetry is provided', () => {
    render(<ErrorState error={null} onRetry={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument()
  })

  it('does not render retry button when onRetry is absent', () => {
    render(<ErrorState error={null} />)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('calls onRetry when retry button is clicked', () => {
    const onRetry = vi.fn()
    render(<ErrorState error={null} onRetry={onRetry} />)
    fireEvent.click(screen.getByRole('button', { name: 'Retry' }))
    expect(onRetry).toHaveBeenCalledOnce()
  })
})
