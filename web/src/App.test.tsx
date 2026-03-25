import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'

import { App } from './App'

// Prevent WebSocket from being instantiated in tests
vi.mock('./api/websocket', () => ({
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(() => vi.fn()),
  },
}))

function renderApp(route = '/') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('App', () => {
  it('renders the Foundry nav brand', () => {
    renderApp()
    expect(screen.getByText('Foundry')).toBeInTheDocument()
  })

  it('renders the project list heading on /', () => {
    renderApp('/')
    expect(screen.getByRole('heading', { name: 'Projects' })).toBeInTheDocument()
  })

  it('renders the New Project button on /', () => {
    renderApp('/')
    expect(screen.getByRole('button', { name: /new project/i })).toBeInTheDocument()
  })

  it('renders the project dashboard on /projects/:id', () => {
    renderApp('/projects/123')
    // Dashboard shows a loading state — just verify we are not on the project list
    expect(screen.queryByRole('button', { name: /new project/i })).not.toBeInTheDocument()
  })

  it('renders the agent detail on /projects/:id/agents/:agentId', () => {
    renderApp('/projects/123/agents/456')
    // Agent detail shows a loading state — verify we are not on the project list
    expect(screen.queryByRole('button', { name: /new project/i })).not.toBeInTheDocument()
  })

  it('renders the Settings nav link when on a project page', () => {
    renderApp('/projects/123')
    expect(screen.getByRole('link', { name: /settings/i })).toBeInTheDocument()
  })
})
