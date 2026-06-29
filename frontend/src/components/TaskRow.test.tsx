import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { TaskRow } from './TaskRow';

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task title',
    body: 'First paragraph\n\nSecond paragraph',
    status: 'todo',
    priority: 2,
    due_date: '2026-07-03',
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

describe('TaskRow', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-01T00:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders status and priority badges', () => {
    render(<TaskRow todo={makeTodo({ status: 'in_progress', priority: 3 })} onSelect={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('in progress')).toBeInTheDocument();
    expect(screen.getByText('High')).toBeInTheDocument();
  });

  it('renders due date as relative text and hides it when null', () => {
    const { rerender } = render(<TaskRow todo={makeTodo()} onSelect={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('Due in 2 days')).toBeInTheDocument();

    rerender(<TaskRow todo={makeTodo({ due_date: null })} onSelect={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.queryByText(/Due/i)).not.toBeInTheDocument();
  });

  it('renders only the first markdown paragraph', () => {
    render(<TaskRow todo={makeTodo()} onSelect={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('First paragraph')).toBeInTheDocument();
    expect(screen.queryByText('Second paragraph')).not.toBeInTheDocument();
  });

  it('calls onSelect and onDelete', () => {
    const onSelect = vi.fn();
    const onDelete = vi.fn();

    render(<TaskRow todo={makeTodo()} onSelect={onSelect} onDelete={onDelete} />);

    const row = screen.getByTestId('task-row-1');
    fireEvent.click(within(row).getByRole('button', { name: /^task title/i }));
    fireEvent.click(within(row).getByRole('button', { name: /delete task title/i }));

    expect(onSelect).toHaveBeenCalledWith(1);
    expect(onDelete).toHaveBeenCalledWith(1);
  });
});
