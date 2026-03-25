import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'

import {
  getProjectsByIdOptions,
  getProjectsByIdRiskProfileOptions,
} from '@/api/generated/@tanstack/react-query.gen'
import { ModelRoutingEditor } from '@/features/settings/ModelRoutingEditor'
import { RiskProfileEditor } from '@/features/settings/RiskProfileEditor'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

export function ProjectSettings() {
  const { id } = useParams<{ id: string }>()
  const projectId = id!

  const { data: project } = useQuery(
    getProjectsByIdOptions({ path: { id: projectId } })
  )

  const { data: riskProfile, isLoading } = useQuery(
    getProjectsByIdRiskProfileOptions({ path: { id: projectId } })
  )

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon-sm" asChild>
          <Link to={`/projects/${projectId}`}>
            <ArrowLeft />
          </Link>
        </Button>
        <h1 className="text-2xl font-bold">
          {project?.name ? `${project.name} — Settings` : 'Settings'}
        </h1>
      </div>

      {isLoading && (
        <div className="text-sm text-muted-foreground">Loading settings...</div>
      )}

      {!isLoading && riskProfile && (
        <Tabs defaultValue="risk">
          <TabsList>
            <TabsTrigger value="risk">Risk Profile</TabsTrigger>
            <TabsTrigger value="routing">Model Routing</TabsTrigger>
          </TabsList>

          <TabsContent value="risk" className="mt-4">
            <RiskProfileEditor
              projectId={projectId}
              riskProfile={riskProfile}
            />
          </TabsContent>

          <TabsContent value="routing" className="mt-4">
            <ModelRoutingEditor
              projectId={projectId}
              riskProfile={riskProfile}
            />
          </TabsContent>
        </Tabs>
      )}

      {!isLoading && !riskProfile && (
        <div className="text-sm text-muted-foreground">
          No risk profile found for this project.
        </div>
      )}
    </div>
  )
}
