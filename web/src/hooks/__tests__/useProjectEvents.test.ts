import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'

vi.mock('@/api/websocket', () => ({
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
  },
}))

vi.mock('@/api/generated/@tanstack/react-query.gen', () => ({
  getProjectsByIdTasksQueryKey: (opts: { path: { id: string } }) => ['tasks', opts.path.id],
}))

import { useProjectEvents } from '../useProjectEvents'
import { wsClient } from '@/api/websocket'

function makeWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useProjectEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls wsClient.connect on mount', () => {
    vi.mocked(wsClient.subscribe).mockReturnValue(vi.fn())
    renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })
    expect(wsClient.connect).toHaveBeenCalled()
  })

  it('calls wsClient.subscribe on mount', () => {
    vi.mocked(wsClient.subscribe).mockReturnValue(vi.fn())
    renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })
    expect(wsClient.subscribe).toHaveBeenCalledWith(expect.any(Function))
  })

  it('calls unsubscribe on unmount', () => {
    const mockUnsubscribe = vi.fn()
    vi.mocked(wsClient.subscribe).mockReturnValue(mockUnsubscribe)
    const { unmount } = renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })
    unmount()
    expect(mockUnsubscribe).toHaveBeenCalled()
  })

  it('does not throw when event type is not task.status_changed', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })

    expect(() => {
      capturedHandler!({ type: 'other.event', project_id: 'proj-1' })
    }).not.toThrow()
  })

  it('ignores events for different projects', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })

    expect(() => {
      capturedHandler!({ type: 'task.status_changed', project_id: 'proj-99' })
    }).not.toThrow()
  })

  it('handles task.status_changed events for the matching project', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    renderHook(() => useProjectEvents('proj-1'), { wrapper: makeWrapper() })

    expect(() => {
      capturedHandler!({ type: 'task.status_changed', project_id: 'proj-1', task_id: 't1' })
    }).not.toThrow()
  })
})
