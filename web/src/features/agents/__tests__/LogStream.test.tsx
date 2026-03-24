import { screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'

import { LogStream } from '../LogStream'
import { renderWithProviders } from '@/test/render'

vi.mock('@/hooks/useAgentLogs', () => ({
  useAgentLogs: vi.fn(),
}))

import { useAgentLogs } from '@/hooks/useAgentLogs'

describe('LogStream', () => {
  it('shows waiting message when no log lines', () => {
    vi.mocked(useAgentLogs).mockReturnValue([])
    renderWithProviders(<LogStream projectId="proj-1" agentId="agent-1" />)
    expect(screen.getByText(/waiting for log output/i)).toBeInTheDocument()
  })

  it('renders each log line', () => {
    vi.mocked(useAgentLogs).mockReturnValue(['Starting agent', 'Processing task', 'Done'])
    renderWithProviders(<LogStream projectId="proj-1" agentId="agent-1" />)
    expect(screen.getByText('Starting agent')).toBeInTheDocument()
    expect(screen.getByText('Processing task')).toBeInTheDocument()
    expect(screen.getByText('Done')).toBeInTheDocument()
  })

  it('renders Agent Logs header', () => {
    vi.mocked(useAgentLogs).mockReturnValue([])
    renderWithProviders(<LogStream projectId="proj-1" agentId="agent-1" />)
    expect(screen.getByText('Agent Logs')).toBeInTheDocument()
  })

  it('passes projectId and agentId to useAgentLogs', () => {
    vi.mocked(useAgentLogs).mockReturnValue([])
    renderWithProviders(<LogStream projectId="proj-42" agentId="agent-99" />)
    expect(useAgentLogs).toHaveBeenCalledWith('proj-42', 'agent-99')
  })
})
