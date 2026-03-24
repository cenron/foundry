import { screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { ProjectList } from '../ProjectList'
import { renderWithProviders } from '@/test/render'
import type { ProjectProject } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsOptions: () => ({ queryKey: ['projects'], queryFn: async () => ({ data: [] }) }),
  postProjectsMutation: () => ({}),
  getProjectsQueryKey: () => ['projects'],
}))

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

const mockProjects: ProjectProject[] = [
  { id: 'p1', name: 'Alpha', status: 'active', description: 'First project' },
  { id: 'p2', name: 'Beta', status: 'draft', description: 'Second project' },
]

describe('ProjectList', () => {
  it('renders the Projects heading', () => {
    vi.mocked(useQuery).mockReturnValue({ data: { data: [] }, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.getByRole('heading', { name: 'Projects' })).toBeInTheDocument()
  })

  it('renders the New Project button', () => {
    vi.mocked(useQuery).mockReturnValue({ data: { data: [] }, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.getAllByRole('button', { name: /new project/i })[0]).toBeInTheDocument()
  })

  it('shows loading state while fetching', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.getByText(/loading projects/i)).toBeInTheDocument()
  })

  it('shows empty state when no projects', () => {
    vi.mocked(useQuery).mockReturnValue({ data: { data: [] }, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.getByText(/no projects yet/i)).toBeInTheDocument()
  })

  it('renders project cards when projects exist', () => {
    vi.mocked(useQuery).mockReturnValue({
      data: { data: mockProjects },
      isLoading: false,
    } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.getByText('Alpha')).toBeInTheDocument()
    expect(screen.getByText('Beta')).toBeInTheDocument()
  })

  it('does not show loading state when data is loaded', () => {
    vi.mocked(useQuery).mockReturnValue({
      data: { data: mockProjects },
      isLoading: false,
    } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.queryByText(/loading projects/i)).not.toBeInTheDocument()
  })

  it('does not show empty state when projects exist', () => {
    vi.mocked(useQuery).mockReturnValue({
      data: { data: mockProjects },
      isLoading: false,
    } as ReturnType<typeof useQuery>)
    renderWithProviders(<ProjectList />)
    expect(screen.queryByText(/no projects yet/i)).not.toBeInTheDocument()
  })
})
