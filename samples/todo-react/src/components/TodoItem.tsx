import type { Todo } from '../types';

interface Props {
  todo: Todo;
  onToggle: (id: string) => void;
  onDelete: (id: string) => void;
}

export function TodoItem({ todo, onToggle, onDelete }: Props) {
  return (
    <li className="flex items-center gap-3 p-4 border-b border-gray-200 group">
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo.id)}
        className="w-5 h-5 rounded-full cursor-pointer accent-green-500"
      />
      <span className={`flex-1 text-lg ${todo.completed ? 'line-through text-gray-400' : 'text-gray-700'}`}>
        {todo.text}
      </span>
      <button
        onClick={() => onDelete(todo.id)}
        className="text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity text-xl leading-none cursor-pointer"
        aria-label="Delete todo"
      >
        ×
      </button>
    </li>
  );
}
