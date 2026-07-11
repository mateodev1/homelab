import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TaskBoard } from '../components/TaskBoard';
import type { Todo } from '../types/todo';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, params, children, ...rest }: Record<string, unknown>) => (
    <a
      href={String(to).replace(
        /\$(\w+)/g,
        (_, key: string) => (params as Record<string, string> | undefined)?.[key] ?? '',
      )}
      {...rest}
    >
      {children as never}
    </a>
  ),
}));

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
    kind: 'note',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

const baseProps = {
  projects: [],
  loading: false,
  error: null,
  activeIndex: 0,
  onSelectTask: vi.fn(),
  onDeleteTask: vi.fn(),
  onStatusChange: vi.fn(),
};

describe('TaskBoard', () => {
  it('renders loading and error states', () => {
    const { rerender } = render(<TaskBoard {...baseProps} tasks={[]} loading={true} />);

    expect(screen.getByLabelText('Loading tasks')).toBeInTheDocument();

    rerender(<TaskBoard {...baseProps} tasks={[]} error="boom" />);
    expect(screen.getByText('Failed to load tasks: boom')).toBeInTheDocument();
  });

  it('shows No tasks when the list is empty', () => {
    render(<TaskBoard {...baseProps} tasks={[]} />);

    expect(screen.getByText('No tasks')).toBeInTheDocument();
  });

  it('renders notes in a grid section and issues in a 4-column board section', () => {
    const onSelectTask = vi.fn();

    render(
      <TaskBoard
        {...baseProps}
        tasks={[
          makeTodo({ id: 1, title: 'A note', kind: 'note', status: 'todo' }),
          makeTodo({ id: 2, title: 'Doing issue', kind: 'issue', status: 'in_progress' }),
          makeTodo({ id: 3, title: 'Done issue', kind: 'issue', status: 'done' }),
          makeTodo({ id: 4, title: 'Cancelled issue', kind: 'issue', status: 'cancelled' }),
        ]}
        activeIndex={1}
        onSelectTask={onSelectTask}
      />,
    );

    expect(screen.getByText('Notes')).toBeInTheDocument();
    expect(screen.getByText('Issues')).toBeInTheDocument();
    expect(screen.getByTestId('note-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-todo')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-in_progress')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-done')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-cancelled')).toBeInTheDocument();
    expect(screen.getByTestId('issue-card-2')).toHaveClass('ring-2');

    fireEvent.click(screen.getByRole('button', { name: 'A note' }));
    expect(onSelectTask).toHaveBeenCalledWith(1);
  });

  it('only renders the Notes section when there are no issues', () => {
    render(
      <TaskBoard {...baseProps} tasks={[makeTodo({ id: 1, title: 'A note', kind: 'note' })]} />,
    );

    expect(screen.getByText('Notes')).toBeInTheDocument();
    expect(screen.queryByText('Issues')).not.toBeInTheDocument();
  });

  it('only renders the Issues section when there are no notes', () => {
    render(
      <TaskBoard {...baseProps} tasks={[makeTodo({ id: 1, title: 'An issue', kind: 'issue' })]} />,
    );

    expect(screen.queryByText('Notes')).not.toBeInTheDocument();
    expect(screen.getByText('Issues')).toBeInTheDocument();
  });
});
