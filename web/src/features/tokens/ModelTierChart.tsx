import type { ApiTaskUsage } from '@/api/generated/types.gen'
import { cn } from '@/lib/utils'

interface ModelTierChartProps {
  breakdown: ApiTaskUsage[]
}

const TIER_CLASSES: Record<string, string> = {
  haiku: 'bg-blue-400',
  sonnet: 'bg-indigo-500',
  opus: 'bg-purple-600',
}

const TIER_LABEL_CLASSES: Record<string, string> = {
  haiku: 'text-blue-700',
  sonnet: 'text-indigo-700',
  opus: 'text-purple-700',
}

export function ModelTierChart({ breakdown }: ModelTierChartProps) {
  const tierCounts: Record<string, number> = {}

  for (const task of breakdown) {
    const tier = task.model_tier ?? 'unknown'
    tierCounts[tier] = (tierCounts[tier] ?? 0) + 1
  }

  const total = Object.values(tierCounts).reduce((a, b) => a + b, 0)
  const tiers = Object.entries(tierCounts).sort((a, b) => b[1] - a[1])

  if (total === 0) {
    return (
      <div className="text-sm text-muted-foreground">No task data yet.</div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex h-4 w-full overflow-hidden rounded-full">
        {tiers.map(([tier, count]) => (
          <div
            key={tier}
            className={cn(TIER_CLASSES[tier] ?? 'bg-gray-400')}
            style={{ width: `${(count / total) * 100}%` }}
            title={`${tier}: ${count} tasks`}
          />
        ))}
      </div>

      <div className="flex flex-wrap gap-3">
        {tiers.map(([tier, count]) => (
          <div key={tier} className="flex items-center gap-1.5 text-xs">
            <span
              className={cn(
                'size-2.5 rounded-sm',
                TIER_CLASSES[tier] ?? 'bg-gray-400'
              )}
            />
            <span
              className={cn(
                'capitalize font-medium',
                TIER_LABEL_CLASSES[tier] ?? 'text-gray-700'
              )}
            >
              {tier}
            </span>
            <span className="text-muted-foreground">
              {count} ({((count / total) * 100).toFixed(0)}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
