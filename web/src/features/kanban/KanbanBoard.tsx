import { useQuery } from '@tanstack/react-query'

import { getProjectsByIdTasksOptions } from '@/api/generated/@tanstack/react-query.gen'
import type { OrchestratorTask } from '@/api/generated/types.gen'
import { KanbanColumn } from './KanbanColumn'

interface KanbanBoardProps {
  projectId: string
}

const BACKLOG_STATUSES = new Set(['pending', 'assigned'])

function groupTasks(tasks: OrchestratorTask[]) {
  return {
    backlog: tasks.filter((t) => BACKLOG_STATUSES.has(t.status ?? '')),
    inProgress: tasks.filter((t) => t.status === 'in_progress'),
    review: tasks.filter((t) => t.status === 'review'),
    done: tasks.filter((t) => t.status === 'completed' || t.status === 'done'),
  }
}

export function KanbanBoard({ projectId }: KanbanBoardProps) {
  const { data: tasks = [], isLoading } = useQuery(
    getProjectsByIdTasksOptions({ path: { id: projectId } })
  )

  if (isLoading) {
    return (
      <div className="text-sm text-muted-foreground">Loading tasks...</div>
    )
  }

  const groups = groupTasks(tasks)

  return (
    <div className="grid grid-cols-4 gap-4">
      <KanbanColumn title="Backlog" tasks={groups.backlog} />
      <KanbanColumn title="In Progress" tasks={groups.inProgress} />
      <KanbanColumn title="Review" tasks={groups.review} />
      <KanbanColumn title="Done" tasks={groups.done} />
    </div>
  )
}
