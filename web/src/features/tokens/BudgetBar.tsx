import { cn } from '@/lib/utils'

interface BudgetBarProps {
  used: number
  budget: number
}

export function BudgetBar({ used, budget }: BudgetBarProps) {
  const pct = budget > 0 ? Math.min((used / budget) * 100, 100) : 0

  const barClass =
    pct > 90
      ? 'bg-red-500'
      : pct > 75
        ? 'bg-yellow-500'
        : 'bg-green-500'

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>{used.toLocaleString()} tokens used</span>
        <span>{pct.toFixed(1)}%</span>
      </div>

      <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={cn('h-full rounded-full transition-all', barClass)}
          style={{ width: `${pct}%` }}
        />
      </div>

      <div className="text-right text-xs text-muted-foreground">
        Budget: {budget.toLocaleString()}
      </div>
    </div>
  )
}
