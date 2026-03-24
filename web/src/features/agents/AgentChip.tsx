import { Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { Pause } from 'lucide-react'

import {
  postProjectsByIdAgentsByAgentIdPauseMutation,
} from '@/api/generated/@tanstack/react-query.gen'
import type { AgentAgent } from '@/api/generated/types.gen'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface AgentChipProps {
  agent: AgentAgent
  projectId: string
  currentTaskTitle?: string
}

const HEALTH_DOT: Record<string, string> = {
  healthy: 'bg-green-500',
  unhealthy: 'bg-yellow-500',
  unresponsive: 'bg-red-500',
}

export function AgentChip({ agent, projectId, currentTaskTitle }: AgentChipProps) {
  const health = agent.health ?? 'unhealthy'
  const dotClass = HEALTH_DOT[health] ?? HEALTH_DOT.unhealthy

  const pauseMutation = useMutation(
    postProjectsByIdAgentsByAgentIdPauseMutation()
  )

  function handlePause(e: React.MouseEvent) {
    e.preventDefault()
    if (!agent.id) return
    pauseMutation.mutate({ path: { id: projectId, agentId: agent.id } })
  }

  return (
    <div className="flex items-center gap-1.5 rounded-lg border border-border bg-card px-2.5 py-1.5 text-sm">
      <span
        className={cn('size-2 shrink-0 rounded-full', dotClass)}
        title={health}
      />

      <Link
        to={`/projects/${projectId}/agents/${agent.id}`}
        className="font-medium hover:underline"
      >
        {agent.role ?? agent.id}
      </Link>

      {currentTaskTitle && (
        <span className="max-w-32 truncate text-xs text-muted-foreground">
          {currentTaskTitle}
        </span>
      )}

      <Button
        variant="ghost"
        size="icon-xs"
        onClick={handlePause}
        disabled={pauseMutation.isPending}
        title="Pause agent"
      >
        <Pause />
      </Button>
    </div>
  )
}
