import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CheckCircle, XCircle } from 'lucide-react'

import {
  getProjectsByIdSpecOptions,
  getProjectsByIdSpecQueryKey,
  postProjectsByIdSpecApproveMutation,
  postProjectsByIdSpecRejectMutation,
} from '@/api/generated/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'

interface SpecViewerProps {
  projectId: string
}

export function SpecViewer({ projectId }: SpecViewerProps) {
  const queryClient = useQueryClient()

  const { data: spec, isLoading } = useQuery(
    getProjectsByIdSpecOptions({ path: { id: projectId } })
  )

  const approveMutation = useMutation({
    ...postProjectsByIdSpecApproveMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getProjectsByIdSpecQueryKey({ path: { id: projectId } }),
      })
    },
  })

  const rejectMutation = useMutation({
    ...postProjectsByIdSpecRejectMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getProjectsByIdSpecQueryKey({ path: { id: projectId } }),
      })
    },
  })

  if (isLoading) {
    return <div className="p-4 text-sm text-muted-foreground">Loading spec...</div>
  }

  if (!spec) {
    return <div className="p-4 text-sm text-muted-foreground">No spec found.</div>
  }

  const isPending = spec.approval_status === 'pending'

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-0.5">
          <span className="text-sm font-medium">Spec</span>
          {spec.token_estimate && (
            <span className="text-xs text-muted-foreground">
              ~{spec.token_estimate.toLocaleString()} tokens
            </span>
          )}
        </div>

        {isPending && (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                rejectMutation.mutate({ path: { id: projectId } })
              }
              disabled={rejectMutation.isPending || approveMutation.isPending}
            >
              <XCircle />
              Reject
            </Button>
            <Button
              size="sm"
              onClick={() =>
                approveMutation.mutate({ path: { id: projectId } })
              }
              disabled={approveMutation.isPending || rejectMutation.isPending}
            >
              <CheckCircle />
              Approve
            </Button>
          </div>
        )}
      </div>

      {spec.approved_content && (
        <section>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Approved
          </h4>
          <pre className="overflow-x-auto rounded-md bg-muted p-3 text-xs leading-relaxed whitespace-pre-wrap">
            {spec.approved_content}
          </pre>
        </section>
      )}

      {spec.execution_content && (
        <section>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Execution Plan
          </h4>
          <pre className="overflow-x-auto rounded-md bg-muted p-3 text-xs leading-relaxed whitespace-pre-wrap">
            {spec.execution_content}
          </pre>
        </section>
      )}
    </div>
  )
}
