import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'

import { KanbanColumn } from '../KanbanColumn'
import { renderWithProviders } from '@/test/render'
import type { OrchestratorTask } from '@/api/generated/types.gen'

const tasks: OrchestratorTask[] = [
  { id: 't1', title: 'First Task', assigned_role: 'developer', risk_level: 'low' },
  { id: 't2', title: 'Second Task', assigned_role: 'reviewer', risk_level: 'high' },
]

describe('KanbanColumn', () => {
  it('renders the column title', () => {
    renderWithProviders(<KanbanColumn title="Backlog" tasks={[]} />)
    expect(screen.getByText('Backlog')).toBeInTheDocument()
  })

  it('shows task count badge with zero', () => {
    renderWithProviders(<KanbanColumn title="Backlog" tasks={[]} />)
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('shows correct task count badge', () => {
    renderWithProviders(<KanbanColumn title="Backlog" tasks={tasks} />)
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('renders all task cards', () => {
    renderWithProviders(<KanbanColumn title="Backlog" tasks={tasks} />)
    expect(screen.getByText('First Task')).toBeInTheDocument()
    expect(screen.getByText('Second Task')).toBeInTheDocument()
  })

  it('renders empty state when no tasks', () => {
    renderWithProviders(<KanbanColumn title="Backlog" tasks={[]} />)
    expect(screen.getByText('No tasks')).toBeInTheDocument()
  })
})
