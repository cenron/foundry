import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { MessageSquare, Send, Square } from 'lucide-react'

import {
  deleteProjectsByIdPoChatMutation,
  postProjectsByIdPoChatMutation,
} from '@/api/generated/@tanstack/react-query.gen'
import { getProjectsByIdPoStatusOptions } from '@/api/generated/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'

interface POChatWindowProps {
  projectId?: string
}

interface SentMessage {
  id: number
  text: string
}

export function POChatWindow({ projectId }: POChatWindowProps) {
  const [messageText, setMessageText] = useState('')
  const [sentMessages, setSentMessages] = useState<SentMessage[]>([])
  const [nextId, setNextId] = useState(1)

  const { data: poStatus } = useQuery({
    ...getProjectsByIdPoStatusOptions({ path: { id: projectId ?? '' } }),
    enabled: !!projectId,
    refetchInterval: 5000,
  })

  const isActive = !!(poStatus as Record<string, unknown> | undefined)?.active

  const sendMutation = useMutation({
    ...postProjectsByIdPoChatMutation(),
    onSuccess: () => {
      setMessageText('')
    },
  })

  const stopMutation = useMutation({
    ...deleteProjectsByIdPoChatMutation(),
  })

  function handleSend() {
    const text = messageText.trim()
    if (!text || !projectId) return

    setSentMessages((prev) => [...prev, { id: nextId, text }])
    setNextId((n) => n + 1)

    sendMutation.mutate({ path: { id: projectId }, body: { message: text } as never })
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      handleSend()
    }
  }

  function handleStop() {
    if (!projectId) return
    stopMutation.mutate({ path: { id: projectId } })
  }

  const isSending = sendMutation.isPending
  const sendDisabled = !messageText.trim() || isSending

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" size="sm">
          <MessageSquare />
          PO Chat
        </Button>
      </SheetTrigger>

      <SheetContent side="right" className="flex flex-col w-96 sm:max-w-96">
        <SheetHeader>
          <SheetTitle>Chat with the PO</SheetTitle>
        </SheetHeader>

        {isActive && (
          <div className="flex items-center justify-between rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
            <span>Session active</span>
            <Button
              size="sm"
              variant="ghost"
              className="h-7 gap-1 text-green-700 hover:text-green-900"
              onClick={handleStop}
              disabled={stopMutation.isPending}
            >
              <Square className="size-3" />
              Stop
            </Button>
          </div>
        )}

        {sendMutation.isError && (
          <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
            Failed to send message. Please try again.
          </div>
        )}

        <div className="flex flex-1 flex-col gap-3 overflow-y-auto">
          {sentMessages.length === 0 && !isActive ? (
            <div className="flex flex-1 flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border p-8 text-center">
              <MessageSquare className="size-8 text-muted-foreground" />
              <p className="text-sm font-medium">No messages yet</p>
              <p className="text-xs text-muted-foreground">
                Send a message to start a PO session.
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-2 pt-2">
              {sentMessages.map((msg) => (
                <div key={msg.id} className="flex justify-end">
                  <div className="max-w-[80%] rounded-lg bg-primary px-3 py-2 text-sm text-primary-foreground">
                    {msg.text}
                  </div>
                </div>
              ))}
              {isSending && (
                <div className="text-xs text-muted-foreground">
                  PO is thinking...
                </div>
              )}
            </div>
          )}
        </div>

        <div className="flex gap-2 border-t border-border pt-4">
          <Input
            placeholder="Message the PO..."
            value={messageText}
            onChange={(e) => setMessageText(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <Button size="icon" aria-label="Send message" disabled={sendDisabled} onClick={handleSend}>
            <Send />
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
