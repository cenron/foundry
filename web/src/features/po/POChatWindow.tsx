import { MessageSquare, Send } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'

export function POChatWindow() {
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
          <SheetTitle>PO Chat</SheetTitle>
        </SheetHeader>

        <div className="flex flex-1 flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border p-8 text-center">
          <MessageSquare className="size-8 text-muted-foreground" />
          <p className="text-sm font-medium">Coming Soon</p>
          <p className="text-xs text-muted-foreground">
            PO integration arrives in Phase 6.
          </p>
        </div>

        <div className="flex gap-2 border-t border-border pt-4">
          <Input placeholder="Message the PO..." disabled />
          <Button size="icon" disabled>
            <Send />
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
