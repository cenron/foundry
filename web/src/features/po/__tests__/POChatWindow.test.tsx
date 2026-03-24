import { screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'

import { POChatWindow } from '../POChatWindow'

describe('POChatWindow', () => {
  it('renders the PO Chat trigger button', () => {
    render(<POChatWindow />)
    expect(screen.getByRole('button', { name: /po chat/i })).toBeInTheDocument()
  })

  it('opens sheet with Coming Soon message on click', async () => {
    render(<POChatWindow />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(await screen.findByText('Coming Soon')).toBeInTheDocument()
  })

  it('shows Phase 6 note inside the sheet', async () => {
    render(<POChatWindow />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(await screen.findByText(/phase 6/i)).toBeInTheDocument()
  })

  it('renders a disabled message input inside the sheet', async () => {
    render(<POChatWindow />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    const input = await screen.findByPlaceholderText(/message the po/i)
    expect(input).toBeDisabled()
  })

  it('renders the sheet heading when open', async () => {
    render(<POChatWindow />)
    fireEvent.click(screen.getByRole('button', { name: /po chat/i }))
    expect(await screen.findByRole('heading', { name: /po chat/i })).toBeInTheDocument()
  })
})
