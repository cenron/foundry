import { Link } from 'react-router-dom'

import type { ProjectProject } from '@/api/generated/types.gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface ProjectCardProps {
  project: ProjectProject
}

const STATUS_CLASSES: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  planning: 'bg-blue-100 text-blue-700',
  active: 'bg-green-100 text-green-700',
  paused: 'bg-yellow-100 text-yellow-700',
  completed: 'bg-purple-100 text-purple-700',
}

export function ProjectCard({ project }: ProjectCardProps) {
  const status = project.status ?? 'draft'
  const statusClass = STATUS_CLASSES[status] ?? STATUS_CLASSES.draft
  const isActive = status === 'active' || status === 'paused'

  return (
    <Card className="flex flex-col justify-between">
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base">{project.name}</CardTitle>
          <Badge
            variant="outline"
            className={cn('shrink-0 capitalize', statusClass)}
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
