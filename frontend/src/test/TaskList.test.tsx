import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TaskList } from '../components/TaskList';
import type { Todo } from '../types/todo';

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

describe('TaskList', () => {
  it('renders loading and error states', () => {
    const { rerender } = render(
      <TaskList
        tasks={[]}
        loading={true}
        error={null}
        activeIndex={0}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );

    expect(screen.getByLabelText('Loading tasks')).toBeInTheDocument();

    rerender(
      <TaskList
        tasks={[]}
        loading={false}
        error="boom"
        activeIndex={0}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );
    expect(screen.getByText('Failed to load tasks: boom')).toBeInTheDocument();
  });

  it('shows No tasks when the list is empty', () => {
    render(
      <TaskList
        tasks={[]}
        loading={false}
        error={null}
        activeIndex={0}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
      />,
    );

    expect(screen.getByText('No tasks')).toBeInTheDocument();
  });

  it('renders every task as a row and highlights the active index', () => {
    const onSelectTask = vi.fn();

    render(
      <TaskList
        tasks={[
          makeTodo({ id: 1, title: 'Todo task', status: 'todo' }),
          makeTodo({ id: 2, title: 'Doing task', status: 'in_progress' }),
          makeTodo({ id: 3, title: 'Done task', status: 'done' }),
          makeTodo({ id: 4, title: 'Cancelled task', status: 'cancelled' }),
        ]}
        loading={false}
        error={null}
        activeIndex={1}
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
    expect(screen.getByTestId('task-row-2')).toHaveClass('bg-accent');
  });
});
