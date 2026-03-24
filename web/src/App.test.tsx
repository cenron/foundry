import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect } from 'vitest'

import { App } from './App'

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

  it('renders the project list on /', () => {
    renderApp('/')
    expect(screen.getByRole('heading', { name: 'Projects' })).toBeInTheDocument()
  })

  it('renders the project dashboard on /projects/:id', () => {
    renderApp('/projects/123')
    expect(screen.getByText('Project Dashboard')).toBeInTheDocument()
  })

  it('renders the agent detail on /projects/:id/agents/:agentId', () => {
    renderApp('/projects/123/agents/456')
    expect(screen.getByText('Agent Detail')).toBeInTheDocument()
  })
})
