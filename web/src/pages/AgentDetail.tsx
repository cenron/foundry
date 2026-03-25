import { useQuery, useMutation } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Pause, Play } from 'lucide-react'

import {
  getProjectsByIdAgentsByAgentIdOptions,
  getProjectsByIdTasksOptions,
  postProjectsByIdAgentsByAgentIdPauseMutation,
  postProjectsByIdAgentsByAgentIdResumeMutation,
} from '@/api/generated/@tanstack/react-query.gen'
import { AgentInfo } from '@/features/agents/AgentInfo'
import { LogStream } from '@/features/agents/LogStream'
import { Button } from '@/components/ui/button'

export function AgentDetail() {
  const { id, agentId } = useParams<{ id: string; agentId: string }>()
  const projectId = id!
  const agentIdParam = agentId!

  const { data: agent, isLoading } = useQuery(
    getProjectsByIdAgentsByAgentIdOptions({
      path: { id: projectId, agentId: agentIdParam },
    })
  )

  const { data: tasks = [] } = useQuery(
    getProjectsByIdTasksOptions({ path: { id: projectId } })
  )

  const pauseMutation = useMutation(
    postProjectsByIdAgentsByAgentIdPauseMutation()
  )

  const resumeMutation = useMutation(
    postProjectsByIdAgentsByAgentIdResumeMutation()
  )

  if (isLoading) {
    return <div className="text-sm text-muted-foreground">Loading agent...</div>
  }

  if (!agent) {
    return <div className="text-sm text-muted-foreground">Agent not found.</div>
  }

  const currentTask = tasks.find((t) => t.id === agent.current_task_id)
  const isPaused = agent.status === 'paused'

  return (
    <div className="flex h-[calc(100vh-8rem)] flex-col gap-4">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon-sm" asChild>
          <Link to={`/projects/${projectId}`}>
            <ArrowLeft />
          </Link>
        </Button>

        <h1 className="text-xl font-bold">{agent.role ?? agentIdParam}</h1>

        <div className="ml-auto flex items-center gap-2">
          {isPaused ? (
            <Button
              size="sm"
              onClick={() =>
                resumeMutation.mutate({
                  path: { id: projectId, agentId: agentIdParam },
                })
              }
              disabled={resumeMutation.isPending}
            >
              <Play />
              Resume
            </Button>
          ) : (
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                pauseMutation.mutate({
                  path: { id: projectId, agentId: agentIdParam },
                })
              }
              disabled={pauseMutation.isPending}
            >
              <Pause />
              Pause
            </Button>
          )}
        </div>
      </div>

      {/* Split layout */}
      <div className="flex flex-1 gap-4 overflow-hidden">
        <div className="w-72 shrink-0 overflow-y-auto">
          <AgentInfo agent={agent} currentTaskTitle={currentTask?.title} />
        </div>

        <div className="flex-1 overflow-hidden">
          <LogStream projectId={projectId} agentId={agentIdParam} />
        </div>
      </div>
    </div>
  )
}
