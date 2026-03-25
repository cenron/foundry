import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useMutation } from '@tanstack/react-query'

import { AgentChip } from '../AgentChip'
import { renderWithProviders } from '@/test/render'
import type { AgentAgent } from '@/api/generated/types.gen'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  postProjectsByIdAgentsByAgentIdPauseMutation: () => ({}),
}))

const mockMutate = vi.fn()

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: mockMutate,
    isPending: false,
  } as ReturnType<typeof useMutation>)
})

const baseAgent: AgentAgent = {
  id: 'agent-1',
  role: 'developer',
  health: 'healthy',
  status: 'running',
}

describe('AgentChip', () => {
  it('renders the agent role', () => {
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    expect(screen.getByText('developer')).toBeInTheDocument()
  })

  it('falls back to agent id when role is absent', () => {
    renderWithProviders(<AgentChip agent={{ ...baseAgent, role: undefined }} projectId="proj-1" />)
    expect(screen.getByText('agent-1')).toBeInTheDocument()
  })

  it('renders a link to the agent detail page', () => {
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    const link = screen.getByRole('link', { name: /developer/i })
    expect(link).toHaveAttribute('href', '/projects/proj-1/agents/agent-1')
  })

  it('renders current task title when provided', () => {
    renderWithProviders(
      <AgentChip agent={baseAgent} projectId="proj-1" currentTaskTitle="Build auth module" />,
    )
    expect(screen.getByText('Build auth module')).toBeInTheDocument()
  })

  it('does not render task title when absent', () => {
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    expect(screen.queryByText(/module/i)).not.toBeInTheDocument()
  })

  it('renders the pause button', () => {
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    expect(screen.getByRole('button', { name: /pause agent/i })).toBeInTheDocument()
  })

  it('calls pause mutate when pause button clicked', () => {
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /pause agent/i }))
    expect(mockMutate).toHaveBeenCalledWith({
      path: { id: 'proj-1', agentId: 'agent-1' },
    })
  })

  it('disables pause button when mutation is pending', () => {
    vi.mocked(useMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as ReturnType<typeof useMutation>)
    renderWithProviders(<AgentChip agent={baseAgent} projectId="proj-1" />)
    expect(screen.getByRole('button', { name: /pause agent/i })).toBeDisabled()
  })

  it('renders unhealthy dot when health is unhealthy', () => {
    renderWithProviders(<AgentChip agent={{ ...baseAgent, health: 'unhealthy' }} projectId="proj-1" />)
    const dot = screen.getByTitle('unhealthy')
    expect(dot).toBeInTheDocument()
  })

  it('renders unresponsive dot when health is unresponsive', () => {
    renderWithProviders(<AgentChip agent={{ ...baseAgent, health: 'unresponsive' }} projectId="proj-1" />)
    expect(screen.getByTitle('unresponsive')).toBeInTheDocument()
  })
})
