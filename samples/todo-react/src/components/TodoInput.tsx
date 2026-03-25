import { useState, type KeyboardEvent } from 'react';

interface Props {
  onAdd: (text: string) => void;
}

export function TodoInput({ onAdd }: Props) {
  const [value, setValue] = useState('');

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' && value.trim()) {
      onAdd(value.trim());
      setValue('');
    }
  }

  return (
    <div className="p-4 border-b border-gray-200">
      <input
        type="text"
        value={value}
        onChange={e => setValue(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="What needs to be done?"
        className="w-full text-lg outline-none text-gray-700 placeholder-gray-400"
        autoFocus
      />
    </div>
  );
}
