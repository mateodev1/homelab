import { cn } from '@/lib/utils';
import { useAuth0 } from '@auth0/auth0-react';
import { CheckSquare, Moon, Plus, Sun } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';
import type { TodoKind, TodoStatus } from '../types/todo';
import { LoginButton } from './LoginButton';
import { LogoutButton } from './LogoutButton';
import { Avatar, AvatarImage } from './ui/avatar';
import { Button } from './ui/button';

export type TaskView = 'all' | TodoStatus | TodoKind;

interface ViewDefinition {
  key: TaskView;
  label: string;
}

export const VIEWS: ViewDefinition[] = [
  { key: 'all', label: 'All' },
  { key: 'todo', label: 'Todo' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'done', label: 'Done' },
  { key: 'cancelled', label: 'Cancelled' },
];

export const KIND_VIEWS: ViewDefinition[] = [
  { key: 'note', label: 'Notes' },
  { key: 'issue', label: 'Issues' },
];

interface SidebarProps {
  activeView: TaskView;
  onSelectView: (view: TaskView) => void;
  counts: Record<TaskView, number>;
  onCreateTask: () => void;
}

export function Sidebar({ activeView, onSelectView, counts, onCreateTask }: SidebarProps) {
  const { theme, toggle } = useTheme();
  const { isAuthenticated, user } = useAuth0();

  return (
    <aside className="flex h-screen w-56 shrink-0 flex-col border-r border-border bg-secondary/40">
      <div className="flex h-12 items-center gap-2 px-3">
        <CheckSquare aria-hidden="true" className="size-4 text-foreground" />
        <span className="text-sm font-semibold tracking-tight text-foreground">Tasks</span>
      </div>

      <div className="px-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCreateTask}
          className="mb-2 w-full justify-between text-sm"
        >
          <span className="flex items-center gap-1.5">
            <Plus aria-hidden="true" className="size-4" />
            New task
          </span>
          <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
            C
          </kbd>
        </Button>
      </div>

      <nav aria-label="Task views" className="flex-1 overflow-y-auto px-2">
        <ul className="grid gap-0.5">
          {VIEWS.map((view) => (
            <li key={view.key}>
              <button
                type="button"
                onClick={() => onSelectView(view.key)}
                aria-current={activeView === view.key ? 'true' : undefined}
                className={cn(
                  'flex w-full items-center justify-between rounded-md px-2 py-1.5 text-sm transition-colors',
                  activeView === view.key
                    ? 'bg-accent text-accent-foreground'
                    : 'text-foreground hover:bg-accent/60',
                )}
              >
                <span>{view.label}</span>
                <span className="text-xs text-muted-foreground">{counts[view.key]}</span>
              </button>
            </li>
          ))}
        </ul>

        <p className="mt-3 px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Type
        </p>
        <ul className="mt-1 grid gap-0.5">
          {KIND_VIEWS.map((view) => (
            <li key={view.key}>
              <button
                type="button"
                onClick={() => onSelectView(view.key)}
                aria-current={activeView === view.key ? 'true' : undefined}
                className={cn(
                  'flex w-full items-center justify-between rounded-md px-2 py-1.5 text-sm transition-colors',
                  activeView === view.key
                    ? 'bg-accent text-accent-foreground'
                    : 'text-foreground hover:bg-accent/60',
                )}
              >
                <span>{view.label}</span>
                <span className="text-xs text-muted-foreground">{counts[view.key]}</span>
              </button>
            </li>
          ))}
        </ul>
      </nav>

      <div className="flex items-center gap-2 border-t border-border px-3 py-2">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={toggle}
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? <Sun aria-hidden="true" /> : <Moon aria-hidden="true" />}
        </Button>

        <div className="flex-1" />

        {isAuthenticated ? (
          <>
            {user?.picture && (
              <Avatar>
                <AvatarImage src={user.picture} alt={user.name ?? 'User'} />
              </Avatar>
            )}
            <LogoutButton />
          </>
        ) : (
          <LoginButton />
        )}
      </div>
    </aside>
  );
}
