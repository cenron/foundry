import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'

import { ModelTierChart } from '../ModelTierChart'
import type { ApiTaskUsage } from '@/api/generated/types.gen'

const breakdown: ApiTaskUsage[] = [
  { task_id: 't1', model_tier: 'haiku', token_usage: 100 },
  { task_id: 't2', model_tier: 'sonnet', token_usage: 200 },
  { task_id: 't3', model_tier: 'sonnet', token_usage: 300 },
  { task_id: 't4', model_tier: 'opus', token_usage: 400 },
]

describe('ModelTierChart', () => {
  it('shows empty state when no breakdown', () => {
    render(<ModelTierChart breakdown={[]} />)
    expect(screen.getByText(/no task data/i)).toBeInTheDocument()
  })

  it('renders tier labels', () => {
    render(<ModelTierChart breakdown={breakdown} />)
    expect(screen.getByText('haiku')).toBeInTheDocument()
    expect(screen.getByText('sonnet')).toBeInTheDocument()
    expect(screen.getByText('opus')).toBeInTheDocument()
  })

  it('renders task counts', () => {
    render(<ModelTierChart breakdown={breakdown} />)
    // haiku: 1 (25%), sonnet: 2 (50%), opus: 1 (25%)
    const fiftyPct = screen.getAllByText(/2 \(50%\)/)
    expect(fiftyPct.length).toBeGreaterThanOrEqual(1)
    const counts = screen.getAllByText(/\d+ \(\d+%\)/)
    expect(counts.length).toBeGreaterThanOrEqual(3)
  })

  it('groups tasks by tier', () => {
    render(<ModelTierChart breakdown={breakdown} />)
    const counts = screen.getAllByText(/\d+ \(\d+%\)/)
    expect(counts).toHaveLength(3)
  })

  it('renders stacked bar segments', () => {
    const { container } = render(<ModelTierChart breakdown={breakdown} />)
    const segments = container.querySelectorAll('[style*="width"]')
    expect(segments.length).toBeGreaterThanOrEqual(3)
  })
})
