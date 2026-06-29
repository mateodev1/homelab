import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { TaskList } from './TaskList';

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

const emptyGrouped = {
  todo: [],
  in_progress: [],
  done: [],
  cancelled: [],
};

describe('TaskList', () => {
  it('renders loading and error states', () => {
    const { rerender } = render(
      <TaskList
        groupedTodos={emptyGrouped}
        loading={true}
        error={null}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );

    expect(screen.getByLabelText('Loading tasks')).toBeInTheDocument();

    rerender(
      <TaskList
        groupedTodos={emptyGrouped}
        loading={false}
        error="boom"
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );
    expect(screen.getByText('Failed to load tasks: boom')).toBeInTheDocument();
  });

  it('always renders every section and shows No tasks when empty', () => {
    render(
      <TaskList
        groupedTodos={emptyGrouped}
        loading={false}
        error={null}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );

    expect(screen.getByText('Todo')).toBeInTheDocument();
    expect(screen.getByText('In Progress')).toBeInTheDocument();
    expect(screen.getByText('Done')).toBeInTheDocument();
    expect(screen.getByText('Cancelled')).toBeInTheDocument();
    expect(screen.getAllByText('No tasks')).toHaveLength(4);
  });

  it('renders tasks in grouped sections', () => {
    const onSelectTask = vi.fn();

    render(
      <TaskList
        groupedTodos={{
          todo: [makeTodo({ id: 1, title: 'Todo task', status: 'todo' })],
          in_progress: [makeTodo({ id: 2, title: 'Doing task', status: 'in_progress' })],
          done: [makeTodo({ id: 3, title: 'Done task', status: 'done' })],
          cancelled: [makeTodo({ id: 4, title: 'Cancelled task', status: 'cancelled' })],
        }}
        loading={false}
        error={null}
        onSelectTask={onSelectTask}
        onDeleteTask={vi.fn()}
      />,
    );

    fireEvent.click(
      within(screen.getByTestId('task-row-1')).getByRole('button', { name: /^todo task/i }),
    );

    expect(screen.getByText('Doing task')).toBeInTheDocument();
    expect(screen.getByText('Done task')).toBeInTheDocument();
    expect(screen.getByText('Cancelled task')).toBeInTheDocument();
    expect(onSelectTask).toHaveBeenCalledWith(1);
  });
});
