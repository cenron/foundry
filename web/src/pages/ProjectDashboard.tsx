import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { Play, Pause, RotateCcw, Settings, FileText, BarChart3 } from 'lucide-react'
import { useState } from 'react'

import {
  getProjectsByIdOptions,
  getProjectsByIdQueryKey,
  postProjectsByIdStartMutation,
  postProjectsByIdPauseMutation,
  postProjectsByIdResumeMutation,
} from '@/api/generated/@tanstack/react-query.gen'
import { AgentStatusBar } from '@/features/agents/AgentStatusBar'
import { KanbanBoard } from '@/features/kanban/KanbanBoard'
import { POChatWindow } from '@/features/po/POChatWindow'
import { SpecViewer } from '@/features/specs/SpecViewer'
import { TokenDashboard } from '@/features/tokens/TokenDashboard'
import { useProjectEvents } from '@/hooks/useProjectEvents'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { statusClass } from '@/lib/status'
import { cn } from '@/lib/utils'

export function ProjectDashboard() {
  const { id } = useParams<{ id: string }>()
  const projectId = id!

  useProjectEvents(projectId)

  const queryClient = useQueryClient()

  const { data: project, isLoading } = useQuery(
    getProjectsByIdOptions({ path: { id: projectId } })
  )

  const invalidateProject = () => {
    queryClient.invalidateQueries({
      queryKey: getProjectsByIdQueryKey({ path: { id: projectId } }),
    })
  }

  const startMutation = useMutation({
    ...postProjectsByIdStartMutation(),
    onSuccess: invalidateProject,
  })

  const pauseMutation = useMutation({
    ...postProjectsByIdPauseMutation(),
    onSuccess: invalidateProject,
  })

  const resumeMutation = useMutation({
    ...postProjectsByIdResumeMutation(),
    onSuccess: invalidateProject,
  })

  const [specOpen, setSpecOpen] = useState(false)
  const [tokensOpen, setTokensOpen] = useState(false)

  if (isLoading) {
    return <div className="text-sm text-muted-foreground">Loading project...</div>
  }

  if (!project) {
    return <div className="text-sm text-muted-foreground">Project not found.</div>
  }

  const status = project.status ?? 'draft'
  const badgeClass = statusClass(status)

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-bold">{project.name}</h1>

        <Badge variant="outline" className={cn('capitalize', badgeClass)}>
          {status}
        </Badge>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          {status === 'draft' || status === 'planning' ? (
            <Button
              size="sm"
              onClick={() => startMutation.mutate({ path: { id: projectId } })}
              disabled={startMutation.isPending}
            >
              <Play />
              Start
            </Button>
          ) : status === 'active' ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => pauseMutation.mutate({ path: { id: projectId } })}
              disabled={pauseMutation.isPending}
            >
              <Pause />
              Pause
            </Button>
          ) : status === 'paused' ? (
            <Button
              size="sm"
              onClick={() => resumeMutation.mutate({ path: { id: projectId } })}
              disabled={resumeMutation.isPending}
            >
              <RotateCcw />
              Resume
            </Button>
          ) : null}

          <POChatWindow />

          {/* Spec panel trigger */}
          <Sheet open={specOpen} onOpenChange={setSpecOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" size="sm">
                <FileText />
                Spec
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[480px] sm:max-w-[480px] overflow-y-auto">
              <SheetHeader>
                <SheetTitle>Project Spec</SheetTitle>
              </SheetHeader>
              <SpecViewer projectId={projectId} />
            </SheetContent>
          </Sheet>

          {/* Token dashboard panel trigger */}
          <Sheet open={tokensOpen} onOpenChange={setTokensOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" size="sm">
                <BarChart3 />
                Tokens
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[480px] sm:max-w-[480px] overflow-y-auto">
              <SheetHeader>
                <SheetTitle>Token Usage</SheetTitle>
              </SheetHeader>
              <TokenDashboard projectId={projectId} />
            </SheetContent>
          </Sheet>

          <Button variant="ghost" size="icon" asChild title="Settings">
            <Link to={`/projects/${projectId}/settings`}>
              <Settings />
            </Link>
          </Button>
        </div>
      </div>

      {/* Agent status bar */}
      <AgentStatusBar projectId={projectId} />

      {/* Kanban board */}
      <KanbanBoard projectId={projectId} />
    </div>
  )
}
