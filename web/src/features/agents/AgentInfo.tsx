import type { AgentAgent } from '@/api/generated/types.gen'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface AgentInfoProps {
  agent: AgentAgent
  currentTaskTitle?: string
}

const HEALTH_CLASSES: Record<string, string> = {
  healthy: 'bg-green-100 text-green-700',
  unhealthy: 'bg-yellow-100 text-yellow-700',
  unresponsive: 'bg-red-100 text-red-700',
}

const STATUS_CLASSES: Record<string, string> = {
  idle: 'bg-gray-100 text-gray-700',
  running: 'bg-green-100 text-green-700',
  paused: 'bg-yellow-100 text-yellow-700',
  stopped: 'bg-red-100 text-red-700',
}

function InfoRow({ label, value }: { label: string; value?: string }) {
  if (!value) return null
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm font-medium break-all">{value}</span>
    </div>
  )
}

export function AgentInfo({ agent, currentTaskTitle }: AgentInfoProps) {
  const health = agent.health ?? 'unhealthy'
  const status = agent.status ?? 'idle'

  return (
    <Card>
      <CardHeader>
        <CardTitle>{agent.role ?? 'Agent'}</CardTitle>
      </CardHeader>

      <CardContent className="flex flex-col gap-3">
        <div className="flex gap-2">
          <Badge
            variant="outline"
            className={cn('capitalize', HEALTH_CLASSES[health] ?? HEALTH_CLASSES.unhealthy)}
          >
            {health}
          </Badge>
          <Badge
            variant="outline"
            className={cn('capitalize', STATUS_CLASSES[status] ?? STATUS_CLASSES.idle)}
          >
            {status}
          </Badge>
        </div>

        <InfoRow label="Provider" value={agent.provider} />
        <InfoRow label="Branch" value={agent.branch_name} />
        <InfoRow label="Worktree" value={agent.worktree_path} />
        <InfoRow label="Current Task" value={currentTaskTitle} />
      </CardContent>
    </Card>
  )
}
