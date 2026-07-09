import { cn } from '@/lib/utils';
import {
  Bug,
  Calendar,
  Circle,
  CircleCheck,
  CircleDot,
  CircleSlash,
  Lightbulb,
  Sparkles,
  StickyNote,
  Trash2,
} from 'lucide-react';
import type { Todo } from '../types/todo';
import { Badge } from './ui/badge';
import { Button } from './ui/button';

const STATUS_ICON: Record<Todo['status'], typeof Circle> = {
  todo: Circle,
  in_progress: CircleDot,
  done: CircleCheck,
  cancelled: CircleSlash,
};

const STATUS_ICON_CLASS: Record<Todo['status'], string> = {
  todo: 'text-muted-foreground',
  in_progress: 'text-blue-500',
  done: 'text-green-600',
  cancelled: 'text-muted-foreground',
};

const STATUS_LABEL: Record<Todo['status'], string> = {
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
  cancelled: 'Cancelled',
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

const KIND_ICON: Record<Todo['kind'], typeof StickyNote> = {
  note: StickyNote,
  issue: Bug,
};

const KIND_ICON_CLASS: Record<Todo['kind'], string> = {
  note: 'text-muted-foreground',
  issue: 'text-orange-500',
};

const KIND_LABEL: Record<Todo['kind'], string> = {
  note: 'Note',
  issue: 'Issue',
};

const ISSUE_TYPE_ICON: Record<NonNullable<Todo['issue_type']>, typeof Bug> = {
  feature: Sparkles,
  bug: Bug,
  improvement: Lightbulb,
};

const ISSUE_TYPE_BADGE_VARIANT: Record<
  NonNullable<Todo['issue_type']>,
  'muted' | 'outline' | 'secondary' | 'destructive'
> = {
  feature: 'secondary',
  bug: 'destructive',
  improvement: 'outline',
};

const ISSUE_TYPE_LABEL: Record<NonNullable<Todo['issue_type']>, string> = {
  feature: 'Feature',
  bug: 'Bug',
  improvement: 'Improvement',
};

interface TaskRowProps {
  todo: Todo;
  active?: boolean;
  onSelect: (id: number) => void;
  onDelete: (id: number) => void;
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

export function TaskRow({ todo, active = false, onSelect, onDelete }: TaskRowProps) {
  const StatusIcon = STATUS_ICON[todo.status];
  const KindIcon = KIND_ICON[todo.kind];
  const IssueTypeIcon = todo.issue_type ? ISSUE_TYPE_ICON[todo.issue_type] : null;

  return (
    <div
      className={cn(
        'group flex items-center gap-3 border-b border-border px-3 py-1.5',
        active ? 'bg-accent' : 'hover:bg-accent/50',
      )}
      data-testid={`task-row-${todo.id}`}
    >
      <StatusIcon
        role="img"
        aria-label={STATUS_LABEL[todo.status]}
        className={cn('size-3.5 shrink-0', STATUS_ICON_CLASS[todo.status])}
      />

      <KindIcon
        role="img"
        aria-label={KIND_LABEL[todo.kind]}
        className={cn('size-3.5 shrink-0', KIND_ICON_CLASS[todo.kind])}
      />

      <button
        type="button"
        className="flex min-w-0 flex-1 items-center gap-3 text-left"
        onClick={() => onSelect(todo.id)}
      >
        <span className="min-w-0 flex-1 truncate text-sm text-foreground">{todo.title}</span>

        {todo.due_date ? (
          <span className="hidden shrink-0 items-center gap-1 text-xs text-muted-foreground sm:flex">
            <Calendar aria-hidden="true" className="size-3" />
            {formatRelativeDate(todo.due_date)}
          </span>
        ) : null}

        {todo.issue_type && IssueTypeIcon ? (
          <Badge variant={ISSUE_TYPE_BADGE_VARIANT[todo.issue_type]} className="shrink-0 gap-1">
            <IssueTypeIcon aria-hidden="true" className="size-3" />
            {ISSUE_TYPE_LABEL[todo.issue_type]}
          </Badge>
        ) : null}

        <Badge variant={PRIORITY_BADGE_VARIANT[todo.priority]} className="shrink-0">
          {priorityLabel(todo.priority)}
        </Badge>
      </button>

      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => onDelete(todo.id)}
        aria-label={`Delete ${todo.title}`}
        className="size-7 shrink-0 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100 focus-visible:opacity-100"
      >
        <Trash2 aria-hidden="true" className="size-3.5" />
      </Button>
    </div>
  );
}
