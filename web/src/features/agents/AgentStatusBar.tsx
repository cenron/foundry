import { useQuery } from '@tanstack/react-query'

import {
  getProjectsByIdAgentsOptions,
  getProjectsByIdTasksOptions,
} from '@/api/generated/@tanstack/react-query.gen'
import type { OrchestratorTask } from '@/api/generated/types.gen'
import { AgentChip } from './AgentChip'

interface AgentStatusBarProps {
  projectId: string
}

function findCurrentTask(tasks: OrchestratorTask[], taskId: string | undefined) {
  if (!taskId) return undefined
  return tasks.find((t) => t.id === taskId)
}

export function AgentStatusBar({ projectId }: AgentStatusBarProps) {
  const { data: agents = [] } = useQuery(
    getProjectsByIdAgentsOptions({ path: { id: projectId } })
  )

  const { data: tasks = [] } = useQuery(
    getProjectsByIdTasksOptions({ path: { id: projectId } })
  )

  if (agents.length === 0) {
    return (
      <div className="text-sm text-muted-foreground">No agents running.</div>
    )
  }

  return (
    <div className="flex flex-wrap gap-2">
      {agents.map((agent) => {
        const currentTask = findCurrentTask(tasks, agent.current_task_id)
        return (
          <AgentChip
            key={agent.id}
            agent={agent}
            projectId={projectId}
            currentTaskTitle={currentTask?.title}
          />
        )
      })}
    </div>
  )
}
