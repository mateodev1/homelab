import {
  Bug,
  Circle,
  CircleCheck,
  CircleDot,
  CircleSlash,
  Lightbulb,
  Sparkles,
} from 'lucide-react';
import type { Todo } from '../types/todo';

export const STATUS_ORDER: Todo['status'][] = ['todo', 'in_progress', 'done', 'cancelled'];

export const STATUS_ICON: Record<Todo['status'], typeof Circle> = {
  todo: Circle,
  in_progress: CircleDot,
  done: CircleCheck,
  cancelled: CircleSlash,
};

export const STATUS_ICON_CLASS: Record<Todo['status'], string> = {
  todo: 'text-muted-foreground',
  in_progress: 'text-blue-500',
  done: 'text-green-600',
  cancelled: 'text-muted-foreground',
};

export const STATUS_LABEL: Record<Todo['status'], string> = {
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
  cancelled: 'Cancelled',
};

export const PRIORITY_BADGE_VARIANT: Record<
  Todo['priority'],
  'muted' | 'outline' | 'secondary' | 'destructive'
> = {
  0: 'muted',
  1: 'outline',
  2: 'secondary',
  3: 'destructive',
};

export const ISSUE_TYPE_ICON: Record<NonNullable<Todo['issue_type']>, typeof Bug> = {
  feature: Sparkles,
  bug: Bug,
  improvement: Lightbulb,
};

export const ISSUE_TYPE_BADGE_VARIANT: Record<
  NonNullable<Todo['issue_type']>,
  'muted' | 'outline' | 'secondary' | 'destructive'
> = {
  feature: 'secondary',
  bug: 'destructive',
  improvement: 'outline',
};

export const ISSUE_TYPE_LABEL: Record<NonNullable<Todo['issue_type']>, string> = {
  feature: 'Feature',
  bug: 'Bug',
  improvement: 'Improvement',
};

export function priorityLabel(priority: Todo['priority']): string {
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

export function formatRelativeDate(dateText: string): string {
  const dueDate = new Date(`${dateText}T00:00:00Z`);
  const now = new Date();

  const msPerDay = 24 * 60 * 60 * 1000;
  const deltaDays = Math.round((dueDate.getTime() - now.getTime()) / msPerDay);
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  return rtf.format(deltaDays, 'day');
}
