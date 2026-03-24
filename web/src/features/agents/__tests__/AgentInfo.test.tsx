import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'

import { AgentInfo } from '../AgentInfo'
import { renderWithProviders } from '@/test/render'
import type { AgentAgent } from '@/api/generated/types.gen'

const baseAgent: AgentAgent = {
  id: 'agent-1',
  role: 'developer',
  health: 'healthy',
  status: 'running',
  provider: 'claude',
  branch_name: 'feat/auth',
  worktree_path: '/workspace/agent-1',
}

describe('AgentInfo', () => {
  it('renders the agent role as title', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('developer')).toBeInTheDocument()
  })

  it('falls back to Agent when role is absent', () => {
    renderWithProviders(<AgentInfo agent={{ ...baseAgent, role: undefined }} />)
    expect(screen.getByText('Agent')).toBeInTheDocument()
  })

  it('renders health badge', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('healthy')).toBeInTheDocument()
  })

  it('renders status badge', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('running')).toBeInTheDocument()
  })

  it('renders provider info row', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('claude')).toBeInTheDocument()
  })

  it('renders branch name info row', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('feat/auth')).toBeInTheDocument()
  })

  it('renders worktree path info row', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} />)
    expect(screen.getByText('/workspace/agent-1')).toBeInTheDocument()
  })

  it('renders current task title when provided', () => {
    renderWithProviders(<AgentInfo agent={baseAgent} currentTaskTitle="Build login" />)
    expect(screen.getByText('Build login')).toBeInTheDocument()
  })

  it('does not render provider row when absent', () => {
    renderWithProviders(<AgentInfo agent={{ ...baseAgent, provider: undefined }} />)
    expect(screen.queryByText('Provider')).not.toBeInTheDocument()
  })

  it('renders unhealthy badge', () => {
    renderWithProviders(<AgentInfo agent={{ ...baseAgent, health: 'unhealthy' }} />)
    expect(screen.getByText('unhealthy')).toBeInTheDocument()
  })

  it('defaults to idle status when status is absent', () => {
    renderWithProviders(<AgentInfo agent={{ ...baseAgent, status: undefined }} />)
    expect(screen.getByText('idle')).toBeInTheDocument()
  })
})
