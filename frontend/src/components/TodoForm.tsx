import { FormEvent, useState } from 'react';

interface TodoFormProps {
  onAdd: (title: string) => void;
}

export function TodoForm({ onAdd }: TodoFormProps) {
  const [title, setTitle] = useState('');

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      return;
    }

    onAdd(trimmedTitle);
    setTitle('');
  };

  return (
    <form aria-label="Add todo form" onSubmit={handleSubmit}>
      <input
        aria-label="Todo title"
        type="text"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
      />
      <button type="submit" disabled={!title.trim()}>
        Add
      </button>
    </form>
  );
}
