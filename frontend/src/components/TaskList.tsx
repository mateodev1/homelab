import { Loader2 } from 'lucide-react';
import type { Todo } from '../types/todo';
import { TaskRow } from './TaskRow';

interface TaskListProps {
  tasks: Todo[];
  loading: boolean;
  error: string | null;
  activeIndex: number;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
}

export function TaskList({
  tasks,
  loading,
  error,
  activeIndex,
  onSelectTask,
  onDeleteTask,
}: TaskListProps) {
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

  return (
    <div className="border-t border-border">
      {tasks.map((todo, index) => (
        <TaskRow
          key={todo.id}
          todo={todo}
          active={index === activeIndex}
          onSelect={onSelectTask}
          onDelete={onDeleteTask}
        />
      ))}
    </div>
  );
}
