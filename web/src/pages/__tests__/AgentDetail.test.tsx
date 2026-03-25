import { screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { AgentDetail } from '../AgentDetail'
import { renderWithRoute } from '@/test/render'
import type { AgentAgent } from '@/api/generated/types.gen'

vi.mock('@/api/websocket', () => ({
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(() => vi.fn()),
  },
}))

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdAgentsByAgentIdOptions: () => ({ queryKey: ['agent'], queryFn: async () => null }),
  getProjectsByIdTasksOptions: () => ({ queryKey: ['tasks'], queryFn: async () => [] }),
  postProjectsByIdAgentsByAgentIdPauseMutation: () => ({}),
  postProjectsByIdAgentsByAgentIdResumeMutation: () => ({}),
}))

const mockAgent: AgentAgent = {
  id: 'agent-1',
  role: 'developer',
  health: 'healthy',
  status: 'running',
  provider: 'claude',
  branch_name: 'feat/login',
}

const routeConfig = {
  pattern: '/projects/:id/agents/:agentId',
  route: '/projects/proj-1/agents/agent-1',
}

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

describe('AgentDetail', () => {
  it('shows loading state while fetching agent', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByText(/loading agent/i)).toBeInTheDocument()
  })

  it('shows not found when agent is null', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: null, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByText(/agent not found/i)).toBeInTheDocument()
  })

  it('renders agent role as heading', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgent, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByRole('heading', { name: 'developer' })).toBeInTheDocument()
  })

  it('renders Pause button for running agent', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgent, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByRole('button', { name: /pause/i })).toBeInTheDocument()
  })

  it('renders Resume button for paused agent', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: { ...mockAgent, status: 'paused' }, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByRole('button', { name: /resume/i })).toBeInTheDocument()
  })

  it('renders back navigation link to project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgent, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    const backLink = screen.getByRole('link')
    expect(backLink).toHaveAttribute('href', '/projects/proj-1')
  })

  it('renders AgentInfo panel with provider', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgent, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByText('claude')).toBeInTheDocument()
  })

  it('renders LogStream area', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockAgent, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<AgentDetail />, routeConfig)
    expect(screen.getByText('Agent Logs')).toBeInTheDocument()
  })
})
