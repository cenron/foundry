import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'

import { BudgetBar } from '../BudgetBar'

describe('BudgetBar', () => {
  it('renders tokens used text', () => {
    render(<BudgetBar used={250000} budget={1000000} />)
    expect(screen.getByText('250,000 tokens used')).toBeInTheDocument()
  })

  it('renders budget text', () => {
    render(<BudgetBar used={250000} budget={1000000} />)
    expect(screen.getByText('Budget: 1,000,000')).toBeInTheDocument()
  })

  it('renders percentage text', () => {
    render(<BudgetBar used={250000} budget={1000000} />)
    expect(screen.getByText('25.0%')).toBeInTheDocument()
  })

  it('renders 0% when budget is zero', () => {
    render(<BudgetBar used={0} budget={0} />)
    expect(screen.getByText('0.0%')).toBeInTheDocument()
  })

  it('caps percentage at 100% when used exceeds budget', () => {
    render(<BudgetBar used={2000000} budget={1000000} />)
    expect(screen.getByText('100.0%')).toBeInTheDocument()
  })

  it('renders progress bar element', () => {
    const { container } = render(<BudgetBar used={500000} budget={1000000} />)
    const bar = container.querySelector('[style*="width"]') as HTMLElement
    expect(bar).toBeTruthy()
    expect(bar.style.width).toBe('50%')
  })

  it('renders green bar when usage is under 75%', () => {
    const { container } = render(<BudgetBar used={500000} budget={1000000} />)
    const bar = container.querySelector('[style*="width"]') as HTMLElement
    expect(bar.className).toContain('bg-green-500')
  })

  it('renders yellow bar when usage is between 75% and 90%', () => {
    const { container } = render(<BudgetBar used={800000} budget={1000000} />)
    const bar = container.querySelector('[style*="width"]') as HTMLElement
    expect(bar.className).toContain('bg-yellow-500')
  })

  it('renders red bar when usage exceeds 90%', () => {
    const { container } = render(<BudgetBar used={950000} budget={1000000} />)
    const bar = container.querySelector('[style*="width"]') as HTMLElement
    expect(bar.className).toContain('bg-red-500')
  })
})
