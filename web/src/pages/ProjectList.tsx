import { useQuery } from '@tanstack/react-query'

import { getProjectsOptions } from '@/api/generated/@tanstack/react-query.gen'
import { CreateProjectDialog } from '@/features/projects/CreateProjectDialog'
import { ProjectCard } from '@/features/projects/ProjectCard'

export function ProjectList() {
  const { data, isLoading } = useQuery(getProjectsOptions())

  const projects = data?.data ?? []

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projects</h1>
        <CreateProjectDialog />
      </div>

      {isLoading && (
        <div className="text-sm text-muted-foreground">Loading projects...</div>
      )}

      {!isLoading && projects.length === 0 && (
        <div className="flex flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-border py-16 text-center">
          <p className="text-sm font-medium">No projects yet</p>
          <p className="text-xs text-muted-foreground">
            Create your first project to get started.
          </p>
          <CreateProjectDialog />
        </div>
      )}

      {projects.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  )
}
