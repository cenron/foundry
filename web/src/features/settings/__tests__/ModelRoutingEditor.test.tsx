import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useMutation } from '@tanstack/react-query'

import { ModelRoutingEditor } from '../ModelRoutingEditor'
import { renderWithProviders } from '@/test/render'
import type { ProjectRiskProfile } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  putProjectsByIdRiskProfileMutation: () => ({}),
  getProjectsByIdRiskProfileQueryKey: () => ['risk-profile'],
}))

const mockMutate = vi.fn()

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: mockMutate,
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

const riskProfile: ProjectRiskProfile = {
  id: 'rp-1',
  project_id: 'proj-1',
  name: 'Default',
  model_routing: { low: 'haiku', medium: 'sonnet', high: 'opus' },
}

describe('ModelRoutingEditor', () => {
  it('renders the descriptive text', () => {
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByText(/choose which model tier/i)).toBeInTheDocument()
  })

  it('renders low, medium, high risk level labels', () => {
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByText('low')).toBeInTheDocument()
    expect(screen.getByText('medium')).toBeInTheDocument()
    expect(screen.getByText('high')).toBeInTheDocument()
  })

  it('renders Save Routing button', () => {
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByRole('button', { name: /save routing/i })).toBeInTheDocument()
  })

  it('calls mutate with routing data on save', () => {
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    fireEvent.click(screen.getByRole('button', { name: /save routing/i }))
    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { id: 'proj-1' },
        body: expect.objectContaining({ model_routing: expect.any(Object) }),
      }),
    )
  })

  it('shows Saving... when mutation is pending', () => {
    vi.mocked(useMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as ReturnType<typeof useMutation>)
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByRole('button', { name: /saving/i })).toBeInTheDocument()
  })

  it('disables save button while pending', () => {
    vi.mocked(useMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as ReturnType<typeof useMutation>)
    renderWithProviders(<ModelRoutingEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
  })
})
