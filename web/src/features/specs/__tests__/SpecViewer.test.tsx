import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useQuery, useMutation } from '@tanstack/react-query'

import { SpecViewer } from '../SpecViewer'
import { renderWithProviders } from '@/test/render'
import type { ProjectSpec } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdSpecOptions: () => ({ queryKey: ['spec'], queryFn: async () => null }),
  getProjectsByIdSpecQueryKey: () => ['spec'],
  postProjectsByIdSpecApproveMutation: () => ({}),
  postProjectsByIdSpecRejectMutation: () => ({}),
}))

const mockApproveMutate = vi.fn()
const mockRejectMutate = vi.fn()

beforeEach(() => {
  vi.mocked(useMutation)
    .mockReturnValueOnce({ mutate: mockApproveMutate, isPending: false } as ReturnType<typeof useMutation>)
    .mockReturnValueOnce({ mutate: mockRejectMutate, isPending: false } as ReturnType<typeof useMutation>)
})

const pendingSpec: ProjectSpec = {
  id: 'spec-1',
  approval_status: 'pending',
  approved_content: 'Build a login system',
  execution_content: 'Step 1: Create auth service',
  token_estimate: 5000,
}

describe('SpecViewer', () => {
  it('shows loading state', () => {
    vi.mocked(useQuery).mockReturnValue({ data: undefined, isLoading: true } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByText(/loading spec/i)).toBeInTheDocument()
  })

  it('shows no spec found when spec is null', () => {
    vi.mocked(useQuery).mockReturnValue({ data: null, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByText(/no spec found/i)).toBeInTheDocument()
  })

  it('renders token estimate when spec has one', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByText(/5,000 tokens/i)).toBeInTheDocument()
  })

  it('renders approved content', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByText('Build a login system')).toBeInTheDocument()
  })

  it('renders execution content', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByText('Step 1: Create auth service')).toBeInTheDocument()
  })

  it('renders approve and reject buttons when status is pending', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.getByRole('button', { name: /approve/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /reject/i })).toBeInTheDocument()
  })

  it('does not render approve/reject buttons when status is not pending', () => {
    vi.mocked(useQuery).mockReturnValue({
      data: { ...pendingSpec, approval_status: 'approved' },
      isLoading: false,
    } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    expect(screen.queryByRole('button', { name: /approve/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /reject/i })).not.toBeInTheDocument()
  })

  it('calls approve mutate when approve button clicked', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /approve/i }))
    expect(mockApproveMutate).toHaveBeenCalledWith({ path: { id: 'proj-1' } })
  })

  it('calls reject mutate when reject button clicked', () => {
    vi.mocked(useQuery).mockReturnValue({ data: pendingSpec, isLoading: false } as ReturnType<typeof useQuery>)
    renderWithProviders(<SpecViewer projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /reject/i }))
    expect(mockRejectMutate).toHaveBeenCalledWith({ path: { id: 'proj-1' } })
  })
})
