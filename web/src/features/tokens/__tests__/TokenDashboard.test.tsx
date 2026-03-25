import { screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useQuery } from '@tanstack/react-query'

import { TokenDashboard } from '../TokenDashboard'
import { renderWithProviders } from '@/test/render'
import type { ApiUsageResponse } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdUsageOptions: () => ({ queryKey: ['usage'], queryFn: async () => null }),
}))

const mockUsage: ApiUsageResponse = {
  project_id: 'proj-1',
  total_tokens: 450000,
  task_breakdown: [
    { task_id: 't1', title: 'Auth service', model_tier: 'sonnet', token_usage: 200000 },
    { task_id: 't2', title: 'DB schema', model_tier: 'haiku', token_usage: 250000 },
  ],
}

describe('TokenDashboard', () => {
  it('shows loading state', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.getByText(/loading usage/i)).toBeInTheDocument()
  })

  it('renders Budget section heading', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockUsage, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.getByText('Budget')).toBeInTheDocument()
  })

  it('renders Model Distribution section heading', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockUsage, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.getByText('Model Distribution')).toBeInTheDocument()
  })

  it('renders per-task usage table when breakdown exists', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockUsage, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.getByText('Per-Task Usage')).toBeInTheDocument()
    expect(screen.getByText('Auth service')).toBeInTheDocument()
    expect(screen.getByText('DB schema')).toBeInTheDocument()
  })

  it('renders model tier in per-task table', () => {
    vi.mocked(useQuery).mockReturnValue({ data: mockUsage, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    // Both the chart legend and table rows will show tier names
    expect(screen.getAllByText('sonnet').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('haiku').length).toBeGreaterThanOrEqual(1)
  })

  it('does not render per-task table when no breakdown', () => {
    vi.mocked(useQuery).mockReturnValue({
      data: { ...mockUsage, task_breakdown: [] },
      isLoading: false,
    } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.queryByText('Per-Task Usage')).not.toBeInTheDocument()
  })

  it('renders budget bar with zero tokens when no usage data', () => {
    vi.mocked(useQuery).mockReturnValue({ data: null, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<TokenDashboard projectId="proj-1" />)
    expect(screen.getByText('0 tokens used')).toBeInTheDocument()
  })
})
