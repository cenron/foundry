import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useMutation } from '@tanstack/react-query'

import { RiskProfileEditor } from '../RiskProfileEditor'
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
  low_criteria: { keywords: ['typo'] },
  medium_criteria: { keywords: ['feature'] },
  high_criteria: { keywords: ['auth', 'payment'] },
  model_routing: { low: 'haiku', medium: 'sonnet', high: 'opus' },
}

describe('RiskProfileEditor', () => {
  it('renders low risk criteria label', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByText(/low risk criteria/i)).toBeInTheDocument()
  })

  it('renders medium risk criteria label', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByText(/medium risk criteria/i)).toBeInTheDocument()
  })

  it('renders high risk criteria label', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByText(/high risk criteria/i)).toBeInTheDocument()
  })

  it('renders 3 textareas for the risk levels', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    const textareas = screen.getAllByRole('textbox')
    expect(textareas).toHaveLength(3)
  })

  it('renders Save Risk Profile button', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByRole('button', { name: /save risk profile/i })).toBeInTheDocument()
  })

  it('calls mutate on save', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    fireEvent.click(screen.getByRole('button', { name: /save risk profile/i }))
    expect(mockMutate).toHaveBeenCalled()
  })

  it('shows invalid JSON error when textarea has invalid JSON', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    const textareas = screen.getAllByRole('textbox')
    fireEvent.change(textareas[0], { target: { value: '{ invalid json' } })
    // The error <p> element contains "Invalid JSON"
    const errorMessages = screen.getAllByText(/invalid json/i)
    const errorParagraph = errorMessages.find((el) => el.tagName === 'P')
    expect(errorParagraph).toBeInTheDocument()
  })

  it('disables save button when mutation is pending', () => {
    vi.mocked(useMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as ReturnType<typeof useMutation>)
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
  })

  it('pre-fills textareas with existing criteria JSON', () => {
    renderWithProviders(<RiskProfileEditor projectId="proj-1" riskProfile={riskProfile} />)
    const textareas = screen.getAllByRole('textbox') as HTMLTextAreaElement[]
    expect(textareas[0].value).toContain('typo')
  })
})
