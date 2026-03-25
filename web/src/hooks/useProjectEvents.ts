import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'

import { wsClient } from '@/api/websocket'
import { getProjectsByIdTasksQueryKey } from '@/api/generated/@tanstack/react-query.gen'

interface TaskStatusEvent {
  type: string
  project_id: string
  task_id?: string
}

export function useProjectEvents(projectId: string): void {
  const queryClient = useQueryClient()

  useEffect(() => {
    wsClient.connect()

    const unsubscribe = wsClient.subscribe((data: unknown) => {
      const event = data as TaskStatusEvent

      if (!event?.type || event.project_id !== projectId) return

      if (event.type === 'task.transition') {
        queryClient.invalidateQueries({
          queryKey: getProjectsByIdTasksQueryKey({ path: { id: projectId } }),
        })
      }
    })

    return () => {
      unsubscribe()
    }
  }, [projectId, queryClient])
}
