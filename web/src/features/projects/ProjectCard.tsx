import { Link } from 'react-router-dom'

import type { ProjectProject } from '@/api/generated/types.gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { statusClass } from '@/lib/status'
import { cn } from '@/lib/utils'

interface ProjectCardProps {
  project: ProjectProject
}

export function ProjectCard({ project }: ProjectCardProps) {
  const status = project.status ?? 'draft'
  const isActive = status === 'active' || status === 'paused'

  return (
    <Card className="flex flex-col justify-between">
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base">{project.name}</CardTitle>
          <Badge
            variant="outline"
            className={cn('shrink-0 capitalize', statusClass(status))}
          >
            {status}
          </Badge>
        </div>
        {project.description && (
          <CardDescription>{project.description}</CardDescription>
        )}
      </CardHeader>

      <CardFooter>
        <Button asChild variant={isActive ? 'default' : 'outline'} size="sm">
          <Link to={`/projects/${project.id}`}>
            {isActive ? 'View' : 'Open'}
          </Link>
        </Button>
      </CardFooter>
    </Card>
  )
}
