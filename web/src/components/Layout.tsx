import { Link, useMatch } from 'react-router-dom'

interface LayoutProps {
  children: React.ReactNode
}

export function Layout({ children }: LayoutProps) {
  const projectMatch = useMatch('/projects/:id/*')
  const projectId = projectMatch?.params.id

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-14 items-center">
            <Link to="/" className="text-xl font-bold text-gray-900">
              Foundry
            </Link>
            <div className="flex items-center gap-4">
              <Link to="/" className="text-sm text-gray-600 hover:text-gray-900">
                Projects
              </Link>
              {projectId && (
                <Link
                  to={`/projects/${projectId}/settings`}
                  className="text-sm text-gray-600 hover:text-gray-900"
                >
                  Settings
                </Link>
              )}
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  )
}
