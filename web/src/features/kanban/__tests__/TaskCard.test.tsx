import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'

import { TaskCard } from '../TaskCard'
import { renderWithProviders } from '@/test/render'
import type { OrchestratorTask } from '@/api/generated/types.gen'

const baseTask: OrchestratorTask = {
  id: 'task-1',
  title: 'Implement login',
  assigned_role: 'developer',
  risk_level: 'medium',
  depends_on: [],
}

describe('TaskCard', () => {
  it('renders task title', () => {
    renderWithProviders(<TaskCard task={baseTask} />)
    expect(screen.getByText('Implement login')).toBeInTheDocument()
  })

  it('renders assigned role badge', () => {
    renderWithProviders(<TaskCard task={baseTask} />)
    expect(screen.getByText('developer')).toBeInTheDocument()
  })

  it('renders risk level badge', () => {
    renderWithProviders(<TaskCard task={baseTask} />)
    expect(screen.getByText('medium')).toBeInTheDocument()
  })

  it('defaults to low risk when risk_level is absent', () => {
    renderWithProviders(<TaskCard task={{ ...baseTask, risk_level: undefined }} />)
    expect(screen.getByText('low')).toBeInTheDocument()
  })

  it('does not render dependency count when no dependencies', () => {
    renderWithProviders(<TaskCard task={baseTask} />)
    expect(screen.queryByText(/dep/i)).not.toBeInTheDocument()
  })

  it('renders dependency count for single dependency', () => {
    renderWithProviders(<TaskCard task={{ ...baseTask, depends_on: ['other-task'] }} />)
    expect(screen.getByText('1 dep')).toBeInTheDocument()
  })

  it('renders dependency count for multiple dependencies', () => {
    renderWithProviders(<TaskCard task={{ ...baseTask, depends_on: ['t1', 't2', 't3'] }} />)
    expect(screen.getByText('3 deps')).toBeInTheDocument()
  })

  it('does not render role badge when role is absent', () => {
    renderWithProviders(<TaskCard task={{ ...baseTask, assigned_role: undefined }} />)
    expect(screen.queryByText('developer')).not.toBeInTheDocument()
  })

  it('renders high risk badge', () => {
    renderWithProviders(<TaskCard task={{ ...baseTask, risk_level: 'high' }} />)
    expect(screen.getByText('high')).toBeInTheDocument()
  })
})
