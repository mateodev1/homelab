import type { Todo } from '../types/todo';
import { TaskRow } from './TaskRow';

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

export function TaskList({ groupedTodos, loading, error, onSelectTask, onDeleteTask }: TaskListProps) {
  if (loading) {
    return (
      <div className="task-list__status">
        <div className="spinner" aria-label="Loading tasks" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="task-list__status task-list__status--error">
        <p>Failed to load tasks: {error}</p>
      </div>
    );
  }

  return (
    <div className="task-list">
      {STATUS_SECTIONS.map((section) => {
        const tasks = groupedTodos[section.key];

        return (
          <section key={section.key} className="task-list__section">
            <h2 className="task-list__section-title">{section.label}</h2>
            {tasks.length === 0 ? (
              <p className="task-list__empty">No tasks</p>
            ) : (
              <div className="task-list__rows">
                {tasks.map((todo) => (
                  <TaskRow key={todo.id} todo={todo} onSelect={onSelectTask} onDelete={onDeleteTask} />
                ))}
              </div>
            )}
          </section>
        );
      })}
    </div>
  );
}
