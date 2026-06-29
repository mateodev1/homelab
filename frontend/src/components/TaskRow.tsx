import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { Todo } from '../types/todo';
import './task-tokens.css';

interface TaskRowProps {
  todo: Todo;
  onSelect: (id: number) => void;
  onDelete: (id: number) => void;
}

function firstParagraph(markdown: string): string {
  const trimmed = markdown.trim();
  if (!trimmed) return '';

  const [paragraph] = trimmed.split(/\n\s*\n/);
  return paragraph.trim();
}

function priorityLabel(priority: Todo['priority']): string {
  switch (priority) {
    case 3:
      return 'High';
    case 2:
      return 'Medium';
    case 1:
      return 'Low';
    default:
      return 'None';
  }
}

function formatRelativeDate(dateText: string): string {
  const dueDate = new Date(`${dateText}T00:00:00Z`);
  const now = new Date();

  const msPerDay = 24 * 60 * 60 * 1000;
  const deltaDays = Math.round((dueDate.getTime() - now.getTime()) / msPerDay);
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  return rtf.format(deltaDays, 'day');
}

export function TaskRow({ todo, onSelect, onDelete }: TaskRowProps) {
  const preview = firstParagraph(todo.body);

  return (
    <article className="task-row" data-testid={`task-row-${todo.id}`}>
      <button type="button" className="task-row__content" onClick={() => onSelect(todo.id)}>
        <header className="task-row__header">
          <h3 className="task-row__title">{todo.title}</h3>
          <div className="task-row__chips">
            <span className={`task-chip task-chip--status-${todo.status}`}>
              {todo.status.replace('_', ' ')}
            </span>
            <span className={`task-chip task-chip--priority-${todo.priority}`}>
              {priorityLabel(todo.priority)}
            </span>
          </div>
        </header>

        {todo.due_date ? (
          <p className="task-row__due-date">Due {formatRelativeDate(todo.due_date)}</p>
        ) : null}

        {preview ? (
          <div className="task-row__preview">
            <Markdown remarkPlugins={[remarkGfm]}>{preview}</Markdown>
          </div>
        ) : (
          <p className="task-row__preview task-row__preview--empty">No description</p>
        )}
      </button>

      <button
        type="button"
        className="task-row__delete"
        onClick={() => onDelete(todo.id)}
        aria-label={`Delete ${todo.title}`}
      >
        Delete
      </button>
    </article>
  );
}
