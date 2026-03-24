import { Routes, Route } from 'react-router-dom'

import { Layout } from './components/Layout'
import { AgentDetail } from './pages/AgentDetail'
import { ProjectDashboard } from './pages/ProjectDashboard'
import { ProjectList } from './pages/ProjectList'
import { ProjectSettings } from './pages/ProjectSettings'

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<ProjectList />} />
        <Route path="/projects/:id" element={<ProjectDashboard />} />
        <Route path="/projects/:id/agents/:agentId" element={<AgentDetail />} />
        <Route path="/projects/:id/settings" element={<ProjectSettings />} />
      </Routes>
    </Layout>
  )
}
