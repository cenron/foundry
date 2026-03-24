import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/websocket', () => ({
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
  },
}))

import { useAgentLogs } from '../useAgentLogs'
import { wsClient } from '@/api/websocket'

describe('useAgentLogs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns empty array initially', () => {
    vi.mocked(wsClient.subscribe).mockReturnValue(vi.fn())
    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))
    expect(result.current).toEqual([])
  })

  it('calls wsClient.connect on mount', () => {
    vi.mocked(wsClient.subscribe).mockReturnValue(vi.fn())
    renderHook(() => useAgentLogs('proj-1', 'agent-1'))
    expect(wsClient.connect).toHaveBeenCalled()
  })

  it('calls wsClient.subscribe on mount', () => {
    vi.mocked(wsClient.subscribe).mockReturnValue(vi.fn())
    renderHook(() => useAgentLogs('proj-1', 'agent-1'))
    expect(wsClient.subscribe).toHaveBeenCalledWith(expect.any(Function))
  })

  it('calls unsubscribe on unmount', () => {
    const mockUnsubscribe = vi.fn()
    vi.mocked(wsClient.subscribe).mockReturnValue(mockUnsubscribe)
    const { unmount } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))
    unmount()
    expect(mockUnsubscribe).toHaveBeenCalled()
  })

  it('appends log lines when matching agent.log event is received', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))

    act(() => {
      capturedHandler!({
        type: 'agent.log',
        project_id: 'proj-1',
        agent_id: 'agent-1',
        line: 'Log line 1',
      })
    })

    expect(result.current).toEqual(['Log line 1'])
  })

  it('ignores events for different agents', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))

    act(() => {
      capturedHandler!({
        type: 'agent.log',
        project_id: 'proj-1',
        agent_id: 'agent-99',
        line: 'Should be ignored',
      })
    })

    expect(result.current).toEqual([])
  })

  it('ignores events for different projects', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))

    act(() => {
      capturedHandler!({
        type: 'agent.log',
        project_id: 'proj-99',
        agent_id: 'agent-1',
        line: 'Should be ignored',
      })
    })

    expect(result.current).toEqual([])
  })

  it('ignores events with wrong type', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))

    act(() => {
      capturedHandler!({
        type: 'task.status_changed',
        project_id: 'proj-1',
        agent_id: 'agent-1',
        line: 'Should be ignored',
      })
    })

    expect(result.current).toEqual([])
  })

  it('caps log lines at 500', () => {
    let capturedHandler: ((data: unknown) => void) | null = null
    vi.mocked(wsClient.subscribe).mockImplementation((handler) => {
      capturedHandler = handler
      return vi.fn()
    })

    const { result } = renderHook(() => useAgentLogs('proj-1', 'agent-1'))

    act(() => {
      for (let i = 0; i < 510; i++) {
        capturedHandler!({
          type: 'agent.log',
          project_id: 'proj-1',
          agent_id: 'agent-1',
          line: `Line ${i}`,
        })
      }
    })

    expect(result.current).toHaveLength(500)
    expect(result.current[0]).toBe('Line 10')
    expect(result.current[499]).toBe('Line 509')
  })
})
