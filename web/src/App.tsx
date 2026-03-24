import { Routes, Route } from 'react-router-dom'

import { Layout } from './components/Layout'

function ProjectList() {
  return <h1 className="text-2xl font-bold">Projects</h1>
}

function ProjectDashboard() {
  return <h1 className="text-2xl font-bold">Project Dashboard</h1>
}

function AgentDetail() {
  return <h1 className="text-2xl font-bold">Agent Detail</h1>
}

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<ProjectList />} />
        <Route path="/projects/:id" element={<ProjectDashboard />} />
        <Route path="/projects/:id/agents/:agentId" element={<AgentDetail />} />
      </Routes>
    </Layout>
  )
}
