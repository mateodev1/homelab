import { cn } from '@/lib/utils';
import { useDraggable } from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import { Link } from '@tanstack/react-router';
import { Calendar, Maximize2, Trash2 } from 'lucide-react';
import {
  ISSUE_TYPE_BADGE_VARIANT,
  ISSUE_TYPE_ICON,
  ISSUE_TYPE_LABEL,
  PRIORITY_BADGE_VARIANT,
  formatRelativeDate,
  priorityLabel,
} from '../lib/taskFormatting';
import type { Project } from '../types/project';
import type { Todo } from '../types/todo';
import { Badge } from './ui/badge';
import { Button, buttonVariants } from './ui/button';

interface IssueCardProps {
  todo: Todo;
  project: Project | null;
  active?: boolean;
  onSelect: (id: number) => void;
  onDelete: (id: number) => void;
}

export function IssueCard({ todo, project, active = false, onSelect, onDelete }: IssueCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `issue-${todo.id}`,
    data: { id: todo.id },
  });

  const IssueTypeIcon = todo.issue_type ? ISSUE_TYPE_ICON[todo.issue_type] : null;

  return (
    <div
      ref={setNodeRef}
      {...attributes}
      {...listeners}
      style={{ transform: CSS.Translate.toString(transform) }}
      className={cn(
        'group flex flex-col gap-2 rounded-md border border-border bg-card p-2.5 shadow-sm transition-shadow hover:shadow-md',
        active && 'ring-2 ring-ring',
        isDragging && 'opacity-50',
      )}
      data-testid={`issue-card-${todo.id}`}
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

      <div className="flex flex-wrap items-center gap-1.5">
        {todo.issue_type && IssueTypeIcon ? (
          <Badge variant={ISSUE_TYPE_BADGE_VARIANT[todo.issue_type]} className="gap-1">
            <IssueTypeIcon aria-hidden="true" className="size-3" />
            {ISSUE_TYPE_LABEL[todo.issue_type]}
          </Badge>
        ) : null}

        {todo.priority > 0 ? (
          <Badge variant={PRIORITY_BADGE_VARIANT[todo.priority]}>
            {priorityLabel(todo.priority)}
          </Badge>
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

        {todo.due_date ? (
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar aria-hidden="true" className="size-3" />
            {formatRelativeDate(todo.due_date)}
          </span>
        ) : null}
      </div>
    </div>
  );
}
