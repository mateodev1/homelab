import { Loader2 } from 'lucide-react';
import type { Todo } from '../types/todo';
import { TaskRow } from './TaskRow';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';

interface GroupedTodos {
  todo: Todo[];
  in_progress: Todo[];
  done: Todo[];
  cancelled: Todo[];
}

interface TaskListProps {
  groupedTodos: GroupedTodos;
  loading: boolean;
  error: string | null;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
}

const STATUS_SECTIONS: Array<{ key: keyof GroupedTodos; label: string }> = [
  { key: 'todo', label: 'Todo' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'done', label: 'Done' },
  { key: 'cancelled', label: 'Cancelled' },
];

export function TaskList({
  groupedTodos,
  loading,
  error,
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

  return (
    <div className="grid gap-4">
      {STATUS_SECTIONS.map((section) => {
        const tasks = groupedTodos[section.key];

        return (
          <Card key={section.key}>
            <CardHeader>
              <CardTitle>{section.label}</CardTitle>
            </CardHeader>
            <CardContent>
              {tasks.length === 0 ? (
                <p className="text-sm text-muted-foreground">No tasks</p>
              ) : (
                <div className="grid gap-2">
                  {tasks.map((todo) => (
                    <TaskRow
                      key={todo.id}
                      todo={todo}
                      onSelect={onSelectTask}
                      onDelete={onDeleteTask}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
