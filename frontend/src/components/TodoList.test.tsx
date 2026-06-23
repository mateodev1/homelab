import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { TodoList } from './TodoList';

describe('TodoList', () => {
  const todos: Todo[] = [
    {
      id: 1,
      title: 'First',
      body: '',
      color: 'default',
      pinned: false,
      done: false,
      created_at: '2026-06-21T02:00:00Z',
      updated_at: '2026-06-21T02:00:00Z',
    },
    {
      id: 2,
      title: 'Second',
      body: '',
      color: 'default',
      pinned: false,
      done: true,
      created_at: '2026-06-21T02:10:00Z',
      updated_at: '2026-06-21T02:10:00Z',
    },
  ];

  it('shows loading state', () => {
    render(<TodoList todos={[]} loading error={null} onToggle={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('shows error state', () => {
    render(
      <TodoList todos={[]} loading={false} error="Failed" onToggle={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByText('Error: Failed')).toBeInTheDocument();
  });

  it('shows empty state when there are no todos', () => {
    render(
      <TodoList todos={[]} loading={false} error={null} onToggle={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByText('No todos yet.')).toBeInTheDocument();
  });

  it('renders todo items when todos exist', () => {
    render(
      <TodoList todos={todos} loading={false} error={null} onToggle={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByText('First')).toBeInTheDocument();
    expect(screen.getByText('Second')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /delete/i })).toHaveLength(2);
  });
});
