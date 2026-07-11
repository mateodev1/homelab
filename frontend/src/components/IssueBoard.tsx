import { DndContext, type DragEndEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core';
import { STATUS_LABEL, STATUS_ORDER } from '../lib/taskFormatting';
import type { Project } from '../types/project';
import type { Todo, TodoStatus } from '../types/todo';
import { IssueColumn } from './IssueColumn';

interface IssueBoardProps {
  issues: Todo[];
  projects: Project[];
  activeId: number | null;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
  onStatusChange: (id: number, status: TodoStatus) => void;
}

function isTodoStatus(value: unknown): value is TodoStatus {
  return typeof value === 'string' && (STATUS_ORDER as string[]).includes(value);
}

export function resolveDragStatusChange(
  issues: Todo[],
  event: DragEndEvent,
): { id: number; status: TodoStatus } | null {
  const nextStatus = event.over?.id;
  const issueId = event.active.data.current?.id as number | undefined;

  if (issueId == null || !isTodoStatus(nextStatus)) return null;

  const issue = issues.find((task) => task.id === issueId);
  if (!issue || issue.status === nextStatus) return null;

  return { id: issueId, status: nextStatus };
}

export function IssueBoard({
  issues,
  projects,
  activeId,
  onSelectTask,
  onDeleteTask,
  onStatusChange,
}: IssueBoardProps) {
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }));

  const handleDragEnd = (event: DragEndEvent) => {
    const change = resolveDragStatusChange(issues, event);
    if (change) {
      onStatusChange(change.id, change.status);
    }
  };

  return (
    <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {STATUS_ORDER.map((status) => (
          <IssueColumn
            key={status}
            status={status}
            label={STATUS_LABEL[status]}
            issues={issues.filter((issue) => issue.status === status)}
            projects={projects}
            activeId={activeId}
            onSelectTask={onSelectTask}
            onDeleteTask={onDeleteTask}
          />
        ))}
      </div>
    </DndContext>
  );
}
