import { cn } from '@/lib/utils';
import { useAuth0 } from '@auth0/auth0-react';
import { CheckSquare, Moon, Plus, Sun, Trash2 } from 'lucide-react';
import { type FormEvent, useState } from 'react';
import { useTheme } from '../context/ThemeContext';
import type { Project } from '../types/project';
import type { TodoKind, TodoStatus } from '../types/todo';
import { LoginButton } from './LoginButton';
import { LogoutButton } from './LogoutButton';
import { Avatar, AvatarImage } from './ui/avatar';
import { Button } from './ui/button';
import { Input } from './ui/input';

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
  projects: Project[];
  activeProjectId: number | null;
  onSelectProject: (id: number | null) => void;
  projectCounts: Record<number, number>;
  onCreateProject: (name: string) => void;
  onDeleteProject: (id: number) => void;
}

export function Sidebar({
  activeView,
  onSelectView,
  counts,
  onCreateTask,
  projects,
  activeProjectId,
  onSelectProject,
  projectCounts,
  onCreateProject,
  onDeleteProject,
}: SidebarProps) {
  const { theme, toggle } = useTheme();
  const { isAuthenticated, user } = useAuth0();
  const [newProjectName, setNewProjectName] = useState('');

  const handleCreateProject = (event: FormEvent) => {
    event.preventDefault();
    const name = newProjectName.trim();
    if (!name) return;
    onCreateProject(name);
    setNewProjectName('');
  };

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

        <p className="mt-3 px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Projects
        </p>
        <ul className="mt-1 grid gap-0.5">
          <li>
            <button
              type="button"
              onClick={() => onSelectProject(null)}
              aria-current={activeProjectId === null ? 'true' : undefined}
              className={cn(
                'flex w-full items-center justify-between rounded-md px-2 py-1.5 text-sm transition-colors',
                activeProjectId === null
                  ? 'bg-accent text-accent-foreground'
                  : 'text-foreground hover:bg-accent/60',
              )}
            >
              <span>All projects</span>
            </button>
          </li>
          {projects.map((project) => (
            <li key={project.id} className="group flex items-center">
              <button
                type="button"
                onClick={() => onSelectProject(project.id)}
                aria-current={activeProjectId === project.id ? 'true' : undefined}
                className={cn(
                  'flex min-w-0 flex-1 items-center justify-between gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
                  activeProjectId === project.id
                    ? 'bg-accent text-accent-foreground'
                    : 'text-foreground hover:bg-accent/60',
                )}
              >
                <span className="flex min-w-0 items-center gap-1.5 truncate">
                  <span
                    aria-hidden="true"
                    className="size-2 shrink-0 rounded-full"
                    style={{
                      backgroundColor: project.color === 'default' ? '#94a3b8' : project.color,
                    }}
                  />
                  <span className="truncate">{project.name}</span>
                </span>
                <span className="text-xs text-muted-foreground">
                  {projectCounts[project.id] ?? 0}
                </span>
              </button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => onDeleteProject(project.id)}
                aria-label={`Delete ${project.name}`}
                className="size-6 shrink-0 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100"
              >
                <Trash2 aria-hidden="true" className="size-3" />
              </Button>
            </li>
          ))}
        </ul>
        <form onSubmit={handleCreateProject} className="mt-1 px-2">
          <Input
            type="text"
            placeholder="New project"
            value={newProjectName}
            onChange={(event) => setNewProjectName(event.target.value)}
            aria-label="New project name"
            className="h-7 text-sm"
          />
        </form>
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
