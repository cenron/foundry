import { useState, useEffect } from 'react';
import type { Todo, Filter } from './types';
import { TodoInput } from './components/TodoInput';
import { TodoList } from './components/TodoList';
import { FilterBar } from './components/FilterBar';

const STORAGE_KEY = 'todos';

function loadTodos(): Todo[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export default function App() {
  const [todos, setTodos] = useState<Todo[]>(loadTodos);
  const [filter, setFilter] = useState<Filter>('all');

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(todos));
  }, [todos]);

  function addTodo(text: string) {
    setTodos(prev => [...prev, { id: crypto.randomUUID(), text, completed: false }]);
  }

  function toggleTodo(id: string) {
    setTodos(prev => prev.map(t => t.id === id ? { ...t, completed: !t.completed } : t));
  }

  function deleteTodo(id: string) {
    setTodos(prev => prev.filter(t => t.id !== id));
  }

  function clearCompleted() {
    setTodos(prev => prev.filter(t => !t.completed));
  }

  const filtered = todos.filter(t => {
    if (filter === 'active') return !t.completed;
    if (filter === 'completed') return t.completed;
    return true;
  });

  const activeCount = todos.filter(t => !t.completed).length;
  const hasCompleted = todos.some(t => t.completed);

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center pt-16 px-4">
      <h1 className="text-6xl font-thin text-red-300 mb-8 tracking-widest">todos</h1>
      <div className="w-full max-w-md bg-white shadow-lg rounded">
        <TodoInput onAdd={addTodo} />
        <TodoList todos={filtered} onToggle={toggleTodo} onDelete={deleteTodo} />
        {todos.length > 0 && (
          <FilterBar
            filter={filter}
            activeCount={activeCount}
            hasCompleted={hasCompleted}
            onFilterChange={setFilter}
            onClearCompleted={clearCompleted}
          />
        )}
      </div>
    </div>
  );
}
