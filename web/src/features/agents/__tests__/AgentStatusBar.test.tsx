import { screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { AgentStatusBar } from '../AgentStatusBar'
import { renderWithProviders } from '@/test/render'
import type { AgentAgent, OrchestratorTask } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdAgentsOptions: () => ({ queryKey: ['agents'], queryFn: async () => [] }),
  getProjectsByIdTasksOptions: () => ({ queryKey: ['tasks'], queryFn: async () => [] }),
  postProjectsByIdAgentsByAgentIdPauseMutation: () => ({}),
}))

const mockAgents: AgentAgent[] = [
  { id: 'agent-1', role: 'developer', health: 'healthy', status: 'running' },
  { id: 'agent-2', role: 'reviewer', health: 'unhealthy', status: 'idle', current_task_id: 'task-1' },
]

const mockTasks: OrchestratorTask[] = [
  { id: 'task-1', title: 'Review PR', status: 'review' },
]

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

describe('AgentStatusBar', () => {
  it('shows empty state when no agents', () => {
    vi.mocked(useQuery).mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<AgentStatusBar projectId="proj-1" />)
    expect(screen.getByText(/no agents running/i)).toBeInTheDocument()
  })

  it('renders a chip for each agent', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgents, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<AgentStatusBar projectId="proj-1" />)
    expect(screen.getByText('developer')).toBeInTheDocument()
    expect(screen.getByText('reviewer')).toBeInTheDocument()
  })

  it('passes current task title to the agent chip', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgents, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: mockTasks, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<AgentStatusBar projectId="proj-1" />)
    expect(screen.getByText('Review PR')).toBeInTheDocument()
  })
})
