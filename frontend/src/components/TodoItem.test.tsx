import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { TodoItem } from './TodoItem';

describe('TodoItem', () => {
  const todo: Todo = {
    id: 1,
    title: 'Write TodoItem tests',
    done: false,
    created_at: '2026-06-21T01:00:00Z',
    updated_at: '2026-06-21T01:00:00Z',
  };

  it('renders title and checkbox state', () => {
    render(<TodoItem todo={todo} onToggle={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('Write TodoItem tests')).toBeInTheDocument();
    expect(screen.getByRole('checkbox')).not.toBeChecked();
  });

  it('calls onToggle with todo id when checkbox is clicked', () => {
    const onToggle = vi.fn();

    render(<TodoItem todo={todo} onToggle={onToggle} onDelete={vi.fn()} />);

    fireEvent.click(screen.getByRole('checkbox'));

    expect(onToggle).toHaveBeenCalledWith(1);
  });

  it('calls onDelete with todo id when delete is clicked', () => {
    const onDelete = vi.fn();

    render(<TodoItem todo={todo} onToggle={vi.fn()} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole('button', { name: /delete/i }));

    expect(onDelete).toHaveBeenCalledWith(1);
  });

  it('applies line-through text style when todo is done', () => {
    render(<TodoItem todo={{ ...todo, done: true }} onToggle={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('Write TodoItem tests')).toHaveStyle({
      textDecoration: 'line-through',
    });
    expect(screen.getByRole('checkbox')).toBeChecked();
  });
});
