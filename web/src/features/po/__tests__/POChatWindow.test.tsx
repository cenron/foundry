import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { render } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'

import { POChatWindow } from '../POChatWindow'

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdPoStatusOptions: () => ({
    queryKey: ['po-status'],
    queryFn: async () => ({ active: false }),
  }),
  postProjectsByIdPoChatMutation: () => ({}),
  deleteProjectsByIdPoChatMutation: () => ({}),
}))

function renderWithQuery(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  )
}

describe('POChatWindow', () => {
  it('renders the PO Chat trigger button', () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    expect(screen.getByRole('button', { name: /po chat/i })).toBeInTheDocument()
  })

  it('opens sheet with "Chat with the PO" heading on click', async () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(
      await screen.findByRole('heading', { name: /chat with the po/i })
    ).toBeInTheDocument()
  })

  it('shows instruction text inside the sheet', async () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(
      await screen.findByText(/send a message to start a po session/i)
    ).toBeInTheDocument()
  })

  it('renders an enabled message input inside the sheet', async () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    const input = await screen.findByPlaceholderText(/message the po/i)
    expect(input).not.toBeDisabled()
  })

  it('renders the sheet heading when open', async () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(
      await screen.findByRole('heading', { name: /chat with the po/i })
    ).toBeInTheDocument()
  })

  it('send button is disabled when input is empty', async () => {
    renderWithQuery(<POChatWindow projectId="proj-1" />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    await screen.findByPlaceholderText(/message the po/i)
    expect(screen.getByRole('button', { name: /send message/i })).toBeDisabled()
  })
})
