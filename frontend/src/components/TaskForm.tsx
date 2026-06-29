import { type FormEvent, Suspense, lazy, useEffect, useState } from 'react';
import type { Todo, TodoStatus } from '../types/todo';

const MDEditor = lazy(async () => {
  const mod = await import('@uiw/react-md-editor');
  return { default: mod.default };
});

interface TaskFormProps {
  todo: Todo | null;
  onCreate: (
    title: string,
    body?: string,
    priority?: 0 | 1 | 2 | 3,
    dueDate?: string | null,
  ) => Promise<void>;
  onUpdate: (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'status' | 'priority' | 'due_date'>>,
  ) => Promise<void>;
  onCancelEdit: () => void;
}

export function TaskForm({ todo, onCreate, onUpdate, onCancelEdit }: TaskFormProps) {
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [status, setStatus] = useState<TodoStatus>('todo');
  const [priority, setPriority] = useState<0 | 1 | 2 | 3>(0);
  const [dueDate, setDueDate] = useState<string>('');

  const isEditing = Boolean(todo);

  useEffect(() => {
    if (!todo) {
      setTitle('');
      setBody('');
      setStatus('todo');
      setPriority(0);
      setDueDate('');
      return;
    }

    setTitle(todo.title);
    setBody(todo.body);
    setStatus(todo.status);
    setPriority(todo.priority);
    setDueDate(todo.due_date ?? '');
  }, [todo]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      return;
    }

    if (todo) {
      await onUpdate(todo.id, {
        title: trimmedTitle,
        body,
        status,
        priority,
        due_date: dueDate || null,
      });
      return;
    }

    await onCreate(trimmedTitle, body, priority, dueDate || null);
    setTitle('');
    setBody('');
    setPriority(0);
    setDueDate('');
  };

  return (
    <form className="task-form" onSubmit={handleSubmit}>
      <h2 className="task-form__title">{isEditing ? 'Edit task' : 'Create task'}</h2>

      <input
        className="task-form__input"
        type="text"
        placeholder="Task title"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
      />

      <div className="task-form__row">
        <label className="task-form__label">
          Priority
          <select
            value={priority}
            onChange={(event) => setPriority(Number(event.target.value) as 0 | 1 | 2 | 3)}
          >
            <option value={0}>None</option>
            <option value={1}>Low</option>
            <option value={2}>Medium</option>
            <option value={3}>High</option>
          </select>
        </label>

        <label className="task-form__label">
          Due date
          <input type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} />
        </label>

        <label className="task-form__label">
          Status
          <select value={status} onChange={(event) => setStatus(event.target.value as TodoStatus)}>
            <option value="todo">Todo</option>
            <option value="in_progress">In progress</option>
            <option value="done">Done</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </label>
      </div>

      <Suspense
        fallback={
          <textarea
            className="task-form__fallback"
            value={body}
            onChange={(event) => setBody(event.target.value)}
            rows={8}
          />
        }
      >
        <div data-color-mode="light">
          <MDEditor
            value={body}
            onChange={(value) => setBody(value ?? '')}
            preview="edit"
            height={240}
            textareaProps={{ placeholder: 'Write markdown...' }}
          />
        </div>
      </Suspense>

      <div className="task-form__actions">
        {isEditing ? (
          <button type="button" onClick={onCancelEdit}>
            Cancel
          </button>
        ) : null}
        <button type="submit">{isEditing ? 'Update task' : 'Add task'}</button>
      </div>
    </form>
  );
}
