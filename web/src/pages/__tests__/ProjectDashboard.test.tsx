import { screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { ProjectDashboard } from '../ProjectDashboard'
import { renderWithRoute } from '@/test/render'
import type { ProjectProject } from '@/api/generated/types.gen'

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
  getProjectsByIdOptions: () => ({ queryKey: ['project'], queryFn: async () => null }),
  getProjectsByIdQueryKey: () => ['project'],
  getProjectsByIdAgentsOptions: () => ({ queryKey: ['agents'], queryFn: async () => [] }),
  getProjectsByIdTasksOptions: () => ({ queryKey: ['tasks'], queryFn: async () => [] }),
  getProjectsByIdTasksQueryKey: () => ['tasks'],
  postProjectsByIdStartMutation: () => ({}),
  postProjectsByIdPauseMutation: () => ({}),
  postProjectsByIdResumeMutation: () => ({}),
  postProjectsByIdAgentsByAgentIdPauseMutation: () => ({}),
}))

const mockProject: ProjectProject = {
  id: 'proj-1',
  name: 'Foundry Core',
  status: 'draft',
  description: 'Build the core system',
}

const routeConfig = {
  pattern: '/projects/:id',
  route: '/projects/proj-1',
}

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

describe('ProjectDashboard', () => {
  it('shows loading state while fetching project', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByText(/loading project/i)).toBeInTheDocument()
  })

  it('shows not found state when project is null', () => {
    vi.mocked(useQuery).mockReturnValue({ data: null, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByText(/project not found/i)).toBeInTheDocument()
  })

  it('renders project name when loaded', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('heading', { name: 'Foundry Core' })).toBeInTheDocument()
  })

  it('renders project status badge', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByText('draft')).toBeInTheDocument()
  })

  it('renders Start button for draft project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
  })

  it('renders Pause button for active project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: { ...mockProject, status: 'active' }, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /pause/i })).toBeInTheDocument()
  })

  it('renders Resume button for paused project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: { ...mockProject, status: 'paused' }, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /resume/i })).toBeInTheDocument()
  })

  it('renders PO Chat button', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /po chat/i })).toBeInTheDocument()
  })

  it('renders Spec button', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /^spec$/i })).toBeInTheDocument()
  })

  it('renders Tokens button', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /tokens/i })).toBeInTheDocument()
  })

  it('renders Start button for planning project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: { ...mockProject, status: 'planning' }, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectDashboard />, routeConfig)
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
  })
})
