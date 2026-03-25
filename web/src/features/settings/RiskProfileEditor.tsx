import { useState, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'

import {
  putProjectsByIdRiskProfileMutation,
  getProjectsByIdRiskProfileQueryKey,
} from '@/api/generated/@tanstack/react-query.gen'
import type { ProjectRiskProfile } from '@/api/generated/types.gen'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'

interface RiskProfileEditorProps {
  projectId: string
  riskProfile: ProjectRiskProfile
}

function jsonToText(val: Record<string, unknown> | undefined): string {
  if (!val) return ''
  return JSON.stringify(val, null, 2)
}

function textToJson(text: string): Record<string, unknown> | undefined {
  try {
    return JSON.parse(text) as Record<string, unknown>
  } catch {
    return undefined
  }
}

export function RiskProfileEditor({ projectId, riskProfile }: RiskProfileEditorProps) {
  const queryClient = useQueryClient()

  const [lowText, setLowText] = useState(jsonToText(riskProfile.low_criteria))
  const [mediumText, setMediumText] = useState(jsonToText(riskProfile.medium_criteria))
  const [highText, setHighText] = useState(jsonToText(riskProfile.high_criteria))

  useEffect(() => {
    setLowText(jsonToText(riskProfile.low_criteria))
    setMediumText(jsonToText(riskProfile.medium_criteria))
    setHighText(jsonToText(riskProfile.high_criteria))
  }, [riskProfile])

  const mutation = useMutation({
    ...putProjectsByIdRiskProfileMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getProjectsByIdRiskProfileQueryKey({ path: { id: projectId } }),
      })
    },
  })

  function handleSave() {
    mutation.mutate({
      path: { id: projectId },
      body: {
        name: riskProfile.name,
        low_criteria: textToJson(lowText),
        medium_criteria: textToJson(mediumText),
        high_criteria: textToJson(highText),
        model_routing: riskProfile.model_routing,
      },
    })
  }

  const hasInvalidJson =
    (lowText && !textToJson(lowText)) ||
    (mediumText && !textToJson(mediumText)) ||
    (highText && !textToJson(highText))

  return (
    <div className="flex flex-col gap-4">
      {(['low', 'medium', 'high'] as const).map((level) => {
        const textMap = { low: lowText, medium: mediumText, high: highText }
        const setTextMap = { low: setLowText, medium: setMediumText, high: setHighText }
        const text = textMap[level]
        const setText = setTextMap[level]
        const isInvalid = !!text && !textToJson(text)

        return (
          <div key={level} className="flex flex-col gap-1.5">
            <label className="text-sm font-medium capitalize">{level} risk criteria</label>
            <Textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              rows={5}
              className="font-mono text-xs"
              placeholder="{}"
              aria-invalid={isInvalid}
            />
            {isInvalid && (
              <p className="text-xs text-destructive">Invalid JSON</p>
            )}
          </div>
        )
      })}

      <div>
        <Button
          onClick={handleSave}
          disabled={!!hasInvalidJson || mutation.isPending}
        >
          {mutation.isPending ? 'Saving...' : 'Save Risk Profile'}
        </Button>
      </div>
    </div>
  )
}
