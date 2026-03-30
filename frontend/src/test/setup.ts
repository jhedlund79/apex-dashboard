import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Chart.js — canvas is not available in jsdom
vi.mock('chart.js', () => ({
  Chart: { register: vi.fn() },
  CategoryScale: class {},
  LinearScale: class {},
  PointElement: class {},
  LineElement: class {},
  BarElement: class {},
  Tooltip: class {},
  Filler: class {},
  Legend: class {},
}))

// Mock react-chartjs-2
vi.mock('react-chartjs-2', () => ({
  Line: () => null,
  Bar: () => null,
}))

// Mock react-plaid-link
vi.mock('react-plaid-link', () => ({
  usePlaidLink: () => ({ open: vi.fn(), ready: false }),
}))

// Silence act() warnings about pending state updates
;(globalThis as Record<string, unknown>).IS_REACT_ACT_ENVIRONMENT = true
