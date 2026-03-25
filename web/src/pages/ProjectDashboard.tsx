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
import { cn } from '@/lib/utils'

const STATUS_CLASSES: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  planning: 'bg-blue-100 text-blue-700',
  active: 'bg-green-100 text-green-700',
  paused: 'bg-yellow-100 text-yellow-700',
  completed: 'bg-purple-100 text-purple-700',
}

export function ProjectDashboard() {
  const { id } = useParams<{ id: string }>()
  const projectId = id!

  useProjectEvents(projectId)

  const queryClient = useQueryClient()

  const { data: project, isLoading } = useQuery({
    ...getProjectsByIdOptions({ path: { id: projectId } }),
    refetchInterval: 5000,
  })

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
  const statusClass = STATUS_CLASSES[status] ?? STATUS_CLASSES.draft

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-bold">{project.name}</h1>

        <Badge variant="outline" className={cn('capitalize', statusClass)}>
          {status}
        </Badge>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          {status === 'active' ? (
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
          ) : (
            <Button
              size="sm"
              onClick={() => startMutation.mutate({ path: { id: projectId } })}
              disabled={startMutation.isPending || status === 'completed'}
            >
              <Play />
              Start
            </Button>
          )}

          <POChatWindow projectId={projectId} />

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
