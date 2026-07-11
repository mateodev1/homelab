import { cn } from '@/lib/utils';
import { Link } from '@tanstack/react-router';
import { Calendar, Maximize2, Trash2 } from 'lucide-react';
import { PRIORITY_BADGE_VARIANT, formatRelativeDate, priorityLabel } from '../lib/taskFormatting';
import type { Project } from '../types/project';
import type { Todo } from '../types/todo';
import { Badge } from './ui/badge';
import { Button, buttonVariants } from './ui/button';

interface NoteCardProps {
  todo: Todo;
  project: Project | null;
  active?: boolean;
  onSelect: (id: number) => void;
  onDelete: (id: number) => void;
}

function bodyPreview(body: string): string {
  return body
    .replace(/[#*_`>~-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

export function NoteCard({ todo, project, active = false, onSelect, onDelete }: NoteCardProps) {
  const preview = bodyPreview(todo.body);

  return (
    <div
      className={cn(
        'group flex flex-col gap-2 rounded-lg border border-border bg-card p-3 shadow-sm transition-shadow hover:shadow-md',
        active && 'ring-2 ring-ring',
      )}
      data-testid={`note-card-${todo.id}`}
    >
      <div className="flex items-start justify-between gap-2">
        <button
          type="button"
          className="min-w-0 flex-1 text-left text-sm font-medium text-foreground"
          onClick={() => onSelect(todo.id)}
        >
          {todo.title}
        </button>

        <Link
          to="/todos/$id"
          params={{ id: String(todo.id) }}
          aria-label={`Open ${todo.title}`}
          className={cn(
            buttonVariants({ variant: 'ghost', size: 'icon' }),
            'size-6 shrink-0 text-muted-foreground opacity-0 hover:text-foreground group-hover:opacity-100 focus-visible:opacity-100',
          )}
        >
          <Maximize2 aria-hidden="true" className="size-3.5" />
        </Link>

        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => onDelete(todo.id)}
          aria-label={`Delete ${todo.title}`}
          className="size-6 shrink-0 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100 focus-visible:opacity-100"
        >
          <Trash2 aria-hidden="true" className="size-3.5" />
        </Button>
      </div>

      {preview ? (
        <button
          type="button"
          onClick={() => onSelect(todo.id)}
          className="line-clamp-6 text-left text-sm text-muted-foreground"
        >
          {preview}
        </button>
      ) : null}

      <div className="mt-auto flex flex-wrap items-center gap-1.5 pt-1">
        {todo.due_date ? (
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar aria-hidden="true" className="size-3" />
            {formatRelativeDate(todo.due_date)}
          </span>
        ) : null}

        {project ? (
          <Badge variant="outline" className="gap-1.5 normal-case">
            <span
              aria-hidden="true"
              className="size-2 shrink-0 rounded-full"
              style={{ backgroundColor: project.color === 'default' ? '#94a3b8' : project.color }}
            />
            {project.name}
          </Badge>
        ) : null}

        {todo.priority > 0 ? (
          <Badge variant={PRIORITY_BADGE_VARIANT[todo.priority]}>
            {priorityLabel(todo.priority)}
          </Badge>
        ) : null}
      </div>
    </div>
  );
}
