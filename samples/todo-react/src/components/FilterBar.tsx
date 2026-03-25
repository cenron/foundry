import type { Filter } from '../types';

interface Props {
  filter: Filter;
  activeCount: number;
  hasCompleted: boolean;
  onFilterChange: (filter: Filter) => void;
  onClearCompleted: () => void;
}

const FILTERS: Filter[] = ['all', 'active', 'completed'];

export function FilterBar({ filter, activeCount, hasCompleted, onFilterChange, onClearCompleted }: Props) {
  return (
    <div className="flex items-center justify-between px-4 py-2 text-sm text-gray-500 border-t border-gray-200">
      <span>{activeCount} item{activeCount !== 1 ? 's' : ''} left</span>

      <div className="flex gap-1">
        {FILTERS.map(f => (
          <button
            key={f}
            onClick={() => onFilterChange(f)}
            className={`px-2 py-1 rounded border cursor-pointer capitalize ${
              filter === f
                ? 'border-red-300 text-red-500'
                : 'border-transparent hover:border-gray-300'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      <button
        onClick={onClearCompleted}
        disabled={!hasCompleted}
        className="hover:text-gray-700 disabled:opacity-0 cursor-pointer"
      >
        Clear completed
      </button>
    </div>
  );
}
