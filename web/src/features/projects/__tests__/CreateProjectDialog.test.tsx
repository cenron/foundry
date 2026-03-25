import { screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useMutation } from '@tanstack/react-query'

import { CreateProjectDialog } from '../CreateProjectDialog'
import { renderWithProviders } from '@/test/render'

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query')
  return {
    ...actual,
    useMutation: vi.fn(),
  }
})

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  postProjectsMutation: () => ({}),
  getProjectsQueryKey: () => ['projects'],
}))

const mockMutate = vi.fn()

beforeEach(() => {
  vi.mocked(useMutation).mockReturnValue({
    mutate: mockMutate,
    isPending: false,
    isError: false,
    isSuccess: false,
    data: undefined,
    error: null,
  } as ReturnType<typeof useMutation>)
})

describe('CreateProjectDialog', () => {
  it('renders the New Project trigger button', () => {
    renderWithProviders(<CreateProjectDialog />)
    expect(screen.getByRole('button', { name: /new project/i })).toBeInTheDocument()
  })

  it('opens the dialog when trigger is clicked', async () => {
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /create project/i })).toBeInTheDocument()
    })
  })

  it('renders name, description, and repository URL fields when open', async () => {
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => {
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/description/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/repository url/i)).toBeInTheDocument()
    })
  })

  it('disables the submit button when name is empty', async () => {
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /create project/i })).toBeDisabled()
    })
  })

  it('enables the submit button when name is entered', async () => {
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => screen.getByLabelText(/name/i))
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'New App' } })
    expect(screen.getByRole('button', { name: /create project/i })).not.toBeDisabled()
  })

  it('calls mutate with name on form submit', async () => {
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => screen.getByLabelText(/name/i))
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'New App' } })
    fireEvent.click(screen.getByRole('button', { name: /create project/i }))
    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({ body: expect.objectContaining({ name: 'New App' }) }),
    )
  })

  it('shows Creating... while pending', async () => {
    vi.mocked(useMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as ReturnType<typeof useMutation>)
    renderWithProviders(<CreateProjectDialog />)
    fireEvent.click(screen.getByRole('button', { name: /new project/i }))
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /creating/i })).toBeInTheDocument()
    })
  })
})
