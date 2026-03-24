import { screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { ProjectSettings } from '../ProjectSettings'
import { renderWithRoute } from '@/test/render'
import type { ProjectProject, ProjectRiskProfile } from '@/api/generated/types.gen'

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
  getProjectsByIdRiskProfileOptions: () => ({ queryKey: ['risk-profile'], queryFn: async () => null }),
  putProjectsByIdRiskProfileMutation: () => ({}),
  getProjectsByIdRiskProfileQueryKey: () => ['risk-profile'],
}))

const mockProject: ProjectProject = {
  id: 'proj-1',
  name: 'Foundry Core',
  status: 'active',
}

const mockRiskProfile: ProjectRiskProfile = {
  id: 'rp-1',
  project_id: 'proj-1',
  name: 'Default',
  low_criteria: {},
  medium_criteria: {},
  high_criteria: {},
  model_routing: { low: 'haiku', medium: 'sonnet', high: 'opus' },
}

const routeConfig = {
  pattern: '/projects/:id/settings',
  route: '/projects/proj-1/settings',
}

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

describe('ProjectSettings', () => {
  it('renders fallback Settings heading when project is not loaded', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: undefined, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    expect(screen.getByRole('heading', { name: 'Settings' })).toBeInTheDocument()
  })

  it('renders project name in the heading when loaded', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: mockRiskProfile, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    expect(screen.getByRole('heading', { name: /Foundry Core.*Settings/ })).toBeInTheDocument()
  })

  it('shows loading state while fetching', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: undefined, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    expect(screen.getByText(/loading settings/i)).toBeInTheDocument()
  })

  it('shows no risk profile message when profile is missing', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: null, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    expect(screen.getByText(/no risk profile found/i)).toBeInTheDocument()
  })

  it('renders Risk Profile and Model Routing tabs when profile exists', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: mockRiskProfile, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    expect(screen.getByRole('tab', { name: /risk profile/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /model routing/i })).toBeInTheDocument()
  })

  it('renders back navigation link to the project', () => {
    vi.mocked(useQuery)
      .mockReturnValueOnce({ data: mockProject, isLoading: false } as ReturnType<typeof useQuery>)
      .mockReturnValueOnce({ data: mockRiskProfile, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithRoute(<ProjectSettings />, routeConfig)
    const backLink = screen.getByRole('link')
    expect(backLink).toHaveAttribute('href', '/projects/proj-1')
  })
})
