import type { OrchestratorTask } from '@/api/generated/types.gen'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface TaskCardProps {
  task: OrchestratorTask
}

const RISK_CLASSES: Record<string, string> = {
  low: 'bg-green-100 text-green-700',
  medium: 'bg-yellow-100 text-yellow-700',
  high: 'bg-red-100 text-red-700',
}

export function TaskCard({ task }: TaskCardProps) {
  const riskLevel = task.risk_level ?? 'low'
  const riskClass = RISK_CLASSES[riskLevel] ?? RISK_CLASSES.low
  const dependencyCount = task.depends_on?.length ?? 0

  return (
    <Card size="sm" className="cursor-default">
      <CardHeader>
        <CardTitle className="text-sm leading-snug">{task.title}</CardTitle>
      </CardHeader>

      <CardContent className="flex flex-wrap items-center gap-1.5">
        {task.assigned_role && (
          <Badge variant="outline" className="text-xs capitalize">
            {task.assigned_role}
          </Badge>
        )}

        <Badge
          variant="outline"
          className={cn('text-xs capitalize', riskClass)}
        >
          {riskLevel}
        </Badge>

        {dependencyCount > 0 && (
          <span className="text-xs text-muted-foreground">
            {dependencyCount} dep{dependencyCount !== 1 ? 's' : ''}
          </span>
        )}
      </CardContent>
    </Card>
  )
}
