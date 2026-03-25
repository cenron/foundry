import { useState, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'

import {
  putProjectsByIdRiskProfileMutation,
  getProjectsByIdRiskProfileQueryKey,
} from '@/api/generated/@tanstack/react-query.gen'
import type { ProjectRiskProfile } from '@/api/generated/types.gen'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface ModelRoutingEditorProps {
  projectId: string
  riskProfile: ProjectRiskProfile
}

const MODEL_OPTIONS = ['haiku', 'sonnet', 'opus'] as const
type ModelOption = (typeof MODEL_OPTIONS)[number]

const RISK_LEVELS = ['low', 'medium', 'high'] as const

function getRouting(riskProfile: ProjectRiskProfile): Record<string, ModelOption> {
  const routing = riskProfile.model_routing as Record<string, ModelOption> | undefined
  return {
    low: routing?.low ?? 'haiku',
    medium: routing?.medium ?? 'sonnet',
    high: routing?.high ?? 'opus',
  }
}

export function ModelRoutingEditor({ projectId, riskProfile }: ModelRoutingEditorProps) {
  const queryClient = useQueryClient()
  const [routing, setRouting] = useState(() => getRouting(riskProfile))

  useEffect(() => {
    setRouting(getRouting(riskProfile))
  }, [riskProfile])

  const mutation = useMutation({
    ...putProjectsByIdRiskProfileMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getProjectsByIdRiskProfileQueryKey({ path: { id: projectId } }),
      })
    },
  })

  function setLevel(level: string, model: ModelOption) {
    setRouting((prev) => ({ ...prev, [level]: model }))
  }

  function handleSave() {
    mutation.mutate({
      path: { id: projectId },
      body: {
        name: riskProfile.name,
        low_criteria: riskProfile.low_criteria,
        medium_criteria: riskProfile.medium_criteria,
        high_criteria: riskProfile.high_criteria,
        model_routing: routing,
      },
    })
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm text-muted-foreground">
        Choose which model tier to use for each risk level.
      </p>

      {RISK_LEVELS.map((level) => (
        <div key={level} className="flex items-center gap-4">
          <span className="w-16 text-sm font-medium capitalize">{level}</span>
          <Select
            value={routing[level]}
            onValueChange={(val) => setLevel(level, val as ModelOption)}
          >
            <SelectTrigger className="w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {MODEL_OPTIONS.map((m) => (
                <SelectItem key={m} value={m} className="capitalize">
                  {m}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      ))}

      <div>
        <Button onClick={handleSave} disabled={mutation.isPending}>
          {mutation.isPending ? 'Saving...' : 'Save Routing'}
        </Button>
      </div>
    </div>
  )
}
