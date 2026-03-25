export const STATUS_CLASSES: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  planning: 'bg-blue-100 text-blue-800',
  estimated: 'bg-indigo-100 text-indigo-800',
  approved: 'bg-cyan-100 text-cyan-800',
  active: 'bg-green-100 text-green-800',
  paused: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-purple-100 text-purple-800',
}

export function statusClass(status: string): string {
  return STATUS_CLASSES[status] ?? STATUS_CLASSES.draft
}
