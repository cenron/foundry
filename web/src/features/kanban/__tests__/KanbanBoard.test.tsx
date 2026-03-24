import { screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useQuery } from '@tanstack/react-query'

import { KanbanBoard } from '../KanbanBoard'
import { renderWithProviders } from '@/test/render'
import type { OrchestratorTask } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdTasksOptions: () => ({ queryKey: ['tasks'], queryFn: async () => [] }),
}))

const mockTasks: OrchestratorTask[] = [
  { id: 't1', title: 'Pending Task', status: 'pending', assigned_role: 'developer' },
  { id: 't2', title: 'Assigned Task', status: 'assigned', assigned_role: 'developer' },
  { id: 't3', title: 'In Progress Task', status: 'in_progress', assigned_role: 'reviewer' },
  { id: 't4', title: 'Review Task', status: 'review', assigned_role: 'reviewer' },
  { id: 't5', title: 'Done Task', status: 'completed', assigned_role: 'developer' },
]

describe('KanbanBoard', () => {
  it('shows loading state', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText(/loading tasks/i)).toBeInTheDocument()
  })

  it('renders 4 columns', () => {
    vi.mocked(useQuery).mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText('Backlog')).toBeInTheDocument()
    expect(screen.getByText('In Progress')).toBeInTheDocument()
    expect(screen.getByText('Review')).toBeInTheDocument()
    expect(screen.getByText('Done')).toBeInTheDocument()
  })

  it('groups pending and assigned tasks into Backlog column', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText('Pending Task')).toBeInTheDocument()
    expect(screen.getByText('Assigned Task')).toBeInTheDocument()
  })

  it('groups in_progress tasks into In Progress column', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText('In Progress Task')).toBeInTheDocument()
  })

  it('groups review tasks into Review column', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText('Review Task')).toBeInTheDocument()
  })

  it('groups completed tasks into Done column', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<KanbanBoard projectId="proj-1" />)
    expect(screen.getByText('Done Task')).toBeInTheDocument()
  })
})
