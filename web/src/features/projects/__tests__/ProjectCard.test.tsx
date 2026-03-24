import { screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'

import { ProjectCard } from '../ProjectCard'
import { renderWithProviders } from '@/test/render'
import type { ProjectProject } from '@/api/generated/types.gen'

const baseProject: ProjectProject = {
  id: 'proj-1',
  name: 'My Project',
  description: 'A test project',
  status: 'draft',
}

describe('ProjectCard', () => {
  it('renders the project name', () => {
    renderWithProviders(<ProjectCard project={baseProject} />)
    expect(screen.getByText('My Project')).toBeInTheDocument()
  })

  it('renders the project description', () => {
    renderWithProviders(<ProjectCard project={baseProject} />)
    expect(screen.getByText('A test project')).toBeInTheDocument()
  })

  it('renders status badge', () => {
    renderWithProviders(<ProjectCard project={baseProject} />)
    expect(screen.getByText('draft')).toBeInTheDocument()
  })

  it('renders Open button for non-active statuses', () => {
    renderWithProviders(<ProjectCard project={baseProject} />)
    expect(screen.getByRole('link', { name: /open/i })).toBeInTheDocument()
  })

  it('renders View button for active status', () => {
    renderWithProviders(<ProjectCard project={{ ...baseProject, status: 'active' }} />)
    expect(screen.getByRole('link', { name: /view/i })).toBeInTheDocument()
  })

  it('renders View button for paused status', () => {
    renderWithProviders(<ProjectCard project={{ ...baseProject, status: 'paused' }} />)
    expect(screen.getByRole('link', { name: /view/i })).toBeInTheDocument()
  })

  it('links to the project page', () => {
    renderWithProviders(<ProjectCard project={baseProject} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/projects/proj-1')
  })

  it('does not render description when absent', () => {
    renderWithProviders(<ProjectCard project={{ ...baseProject, description: undefined }} />)
    expect(screen.queryByText('A test project')).not.toBeInTheDocument()
  })

  it('defaults to draft status when status is missing', () => {
    renderWithProviders(<ProjectCard project={{ ...baseProject, status: undefined }} />)
    expect(screen.getByText('draft')).toBeInTheDocument()
  })

  it('renders planning status badge', () => {
    renderWithProviders(<ProjectCard project={{ ...baseProject, status: 'planning' }} />)
    expect(screen.getByText('planning')).toBeInTheDocument()
  })
})
