import { Calendar, Trash2 } from 'lucide-react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { Todo } from '../types/todo';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Card } from './ui/card';

const STATUS_BADGE_VARIANT: Record<Todo['status'], 'muted' | 'default' | 'secondary' | 'outline'> =
  {
    todo: 'outline',
    in_progress: 'secondary',
    done: 'default',
    cancelled: 'muted',
  };

const PRIORITY_BADGE_VARIANT: Record<
  Todo['priority'],
  'muted' | 'outline' | 'secondary' | 'destructive'
> = {
  0: 'muted',
  1: 'outline',
  2: 'secondary',
  3: 'destructive',
};

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
    <Card className="flex flex-row items-start gap-3 p-3" data-testid={`task-row-${todo.id}`}>
      <button type="button" className="flex-1 text-left" onClick={() => onSelect(todo.id)}>
        <header className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-medium text-foreground">{todo.title}</h3>
          <div className="flex shrink-0 gap-1.5">
            <Badge variant={STATUS_BADGE_VARIANT[todo.status]}>
              {todo.status.replace('_', ' ')}
            </Badge>
            <Badge variant={PRIORITY_BADGE_VARIANT[todo.priority]}>
              {priorityLabel(todo.priority)}
            </Badge>
          </div>
        </header>

        {todo.due_date ? (
          <p className="mt-1.5 flex items-center gap-1.5 text-xs text-muted-foreground">
            <Calendar aria-hidden="true" className="size-3.5" />
            Due {formatRelativeDate(todo.due_date)}
          </p>
        ) : null}

        {preview ? (
          <div className="task-row-preview-clamp mt-2 text-sm leading-snug text-foreground">
            <Markdown remarkPlugins={[remarkGfm]}>{preview}</Markdown>
          </div>
        ) : (
          <p className="mt-2 text-sm italic text-muted-foreground">No description</p>
        )}
      </button>

      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => onDelete(todo.id)}
        aria-label={`Delete ${todo.title}`}
        className="shrink-0 text-muted-foreground hover:text-destructive"
      >
        <Trash2 aria-hidden="true" />
      </Button>
    </Card>
  );
}
