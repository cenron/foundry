import { useEffect, useState } from 'react'

import { wsClient } from '@/api/websocket'

const MAX_LOG_LINES = 500

interface LogEvent {
  type: string
  agent_id: string
  project_id: string
  line: string
}

export function useAgentLogs(projectId: string, agentId: string): string[] {
  const [lines, setLines] = useState<string[]>([])

  useEffect(() => {
    wsClient.connect()

    const unsubscribe = wsClient.subscribe((data: unknown) => {
      const event = data as LogEvent

      if (
        event?.type !== 'agent.log' ||
        event.project_id !== projectId ||
        event.agent_id !== agentId
      ) {
        return
      }

      setLines((prev) => {
        const next = [...prev, event.line]
        if (next.length > MAX_LOG_LINES) {
          return next.slice(next.length - MAX_LOG_LINES)
        }
        return next
      })
    })

    return () => {
      unsubscribe()
    }
  }, [projectId, agentId])

  return lines
}
