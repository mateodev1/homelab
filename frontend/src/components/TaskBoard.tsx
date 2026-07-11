import { Loader2 } from 'lucide-react';
import type { Project } from '../types/project';
import type { Todo, TodoStatus } from '../types/todo';
import { IssueBoard } from './IssueBoard';
import { NotesGrid } from './NotesGrid';

interface TaskBoardProps {
  tasks: Todo[];
  projects: Project[];
  loading: boolean;
  error: string | null;
  activeIndex: number;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
  onStatusChange: (id: number, status: TodoStatus) => void;
}

export function TaskBoard({
  tasks,
  projects,
  loading,
  error,
  activeIndex,
  onSelectTask,
  onDeleteTask,
  onStatusChange,
}: TaskBoardProps) {
  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2
          role="status"
          aria-label="Loading tasks"
          className="size-8 animate-spin text-muted-foreground"
        />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center py-8">
        <p className="text-sm text-destructive">Failed to load tasks: {error}</p>
      </div>
    );
  }

  if (tasks.length === 0) {
    return (
      <div className="flex justify-center py-12">
        <p className="text-sm text-muted-foreground">No tasks</p>
      </div>
    );
  }

  const activeId = tasks[activeIndex]?.id ?? null;
  const notes = tasks.filter((task) => task.kind === 'note');
  const issues = tasks.filter((task) => task.kind === 'issue');

  return (
    <div className="flex flex-col gap-6 p-4">
      {notes.length > 0 ? (
        <section>
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Notes
          </h2>
          <NotesGrid
            notes={notes}
            projects={projects}
            activeId={activeId}
            onSelectTask={onSelectTask}
            onDeleteTask={onDeleteTask}
          />
        </section>
      ) : null}

      {issues.length > 0 ? (
        <section>
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Issues
          </h2>
          <IssueBoard
            issues={issues}
            projects={projects}
            activeId={activeId}
            onSelectTask={onSelectTask}
            onDeleteTask={onDeleteTask}
            onStatusChange={onStatusChange}
          />
        </section>
      ) : null}
    </div>
  );
}
