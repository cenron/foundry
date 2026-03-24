import { useQuery } from '@tanstack/react-query'

import { getProjectsByIdUsageOptions } from '@/api/generated/@tanstack/react-query.gen'
import { BudgetBar } from './BudgetBar'
import { ModelTierChart } from './ModelTierChart'

interface TokenDashboardProps {
  projectId: string
}

const DEFAULT_BUDGET = 1_000_000

export function TokenDashboard({ projectId }: TokenDashboardProps) {
  const { data: usage, isLoading } = useQuery(
    getProjectsByIdUsageOptions({ path: { id: projectId } })
  )

  if (isLoading) {
    return <div className="p-4 text-sm text-muted-foreground">Loading usage...</div>
  }

  const totalTokens = usage?.total_tokens ?? 0
  const breakdown = usage?.task_breakdown ?? []

  return (
    <div className="flex flex-col gap-6 p-4">
      <section>
        <h4 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Budget
        </h4>
        <BudgetBar used={totalTokens} budget={DEFAULT_BUDGET} />
      </section>

      <section>
        <h4 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Model Distribution
        </h4>
        <ModelTierChart breakdown={breakdown} />
      </section>

      {breakdown.length > 0 && (
        <section>
          <h4 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Per-Task Usage
          </h4>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border bg-muted/50 text-left">
                  <th className="px-3 py-2 font-medium text-muted-foreground">Task</th>
                  <th className="px-3 py-2 font-medium text-muted-foreground">Tier</th>
                  <th className="px-3 py-2 text-right font-medium text-muted-foreground">Tokens</th>
                </tr>
              </thead>
              <tbody>
                {breakdown.map((row) => (
                  <tr key={row.task_id} className="border-b border-border last:border-0">
                    <td className="px-3 py-2 max-w-48 truncate">{row.title ?? row.task_id}</td>
                    <td className="px-3 py-2 capitalize">{row.model_tier ?? '—'}</td>
                    <td className="px-3 py-2 text-right tabular-nums">
                      {(row.token_usage ?? 0).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  )
}
