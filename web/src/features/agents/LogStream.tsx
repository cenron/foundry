import { useEffect, useRef } from 'react'

import { useAgentLogs } from '@/hooks/useAgentLogs'

interface LogStreamProps {
  projectId: string
  agentId: string
}

export function LogStream({ projectId, agentId }: LogStreamProps) {
  const lines = useAgentLogs(projectId, agentId)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-black text-green-400">
      <div className="border-b border-border/30 px-3 py-2 text-xs text-muted-foreground">
        Agent Logs
      </div>

      <div className="flex-1 overflow-y-auto p-3 font-mono text-xs leading-relaxed">
        {lines.length === 0 ? (
          <span className="text-muted-foreground">Waiting for log output...</span>
        ) : (
          lines.map((line, i) => (
            <div key={i} className="whitespace-pre-wrap break-all">
              {line}
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
