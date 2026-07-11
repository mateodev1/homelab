import { cn } from '@/lib/utils';
import { useDroppable } from '@dnd-kit/core';
import type { Project } from '../types/project';
import type { Todo } from '../types/todo';
import { IssueCard } from './IssueCard';

interface IssueColumnProps {
  status: Todo['status'];
  label: string;
  issues: Todo[];
  projects: Project[];
  activeId: number | null;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
}

export function IssueColumn({
  status,
  label,
  issues,
  projects,
  activeId,
  onSelectTask,
  onDeleteTask,
}: IssueColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id: status });

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'flex min-h-40 flex-col gap-2 rounded-lg border border-border bg-secondary/40 p-2 transition-colors',
        isOver && 'bg-accent/60',
      )}
      data-testid={`issue-column-${status}`}
    >
      <div className="flex items-center justify-between px-1 py-0.5">
        <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          {label}
        </span>
        <span className="text-xs text-muted-foreground">{issues.length}</span>
      </div>

      <div className="flex flex-col gap-2">
        {issues.map((issue) => (
          <IssueCard
            key={issue.id}
            todo={issue}
            project={projects.find((project) => project.id === issue.project_id) ?? null}
            active={issue.id === activeId}
            onSelect={onSelectTask}
            onDelete={onDeleteTask}
          />
        ))}
      </div>
    </div>
  );
}
