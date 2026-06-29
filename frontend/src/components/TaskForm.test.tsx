import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { TaskForm } from './TaskForm';

vi.mock('@uiw/react-md-editor', () => ({
  default: ({ value, onChange }: { value?: string; onChange?: (v: string) => void }) => (
    <textarea
      aria-label="markdown-editor"
      value={value ?? ''}
      onChange={(event) => onChange?.(event.target.value)}
    />
  ),
}));

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 7,
    title: 'Existing task',
    body: 'Existing body',
    status: 'in_progress',
    priority: 2,
    due_date: '2026-07-01',
    created_at: '2026-06-20T10:00:00Z',
    updated_at: '2026-06-20T10:00:00Z',
    ...overrides,
  };
}

describe('TaskForm', () => {
  it('creates a task with markdown body', async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);

    render(<TaskForm todo={null} onCreate={onCreate} onUpdate={vi.fn()} onCancelEdit={vi.fn()} />);

    fireEvent.change(screen.getByPlaceholderText('Task title'), { target: { value: 'New task' } });

    await waitFor(() => {
      expect(screen.getByLabelText('markdown-editor')).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText('markdown-editor'), { target: { value: '# Markdown body' } });

    fireEvent.change(screen.getByLabelText('Priority'), {
      target: { value: '3' },
    });
    fireEvent.change(screen.getByLabelText('Due date'), {
      target: { value: '2026-07-02' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Add task' }));

    await waitFor(() => {
      expect(onCreate).toHaveBeenCalledWith('New task', '# Markdown body', 3, '2026-07-02');
    });
  });

  it('edits an existing task', async () => {
    const onUpdate = vi.fn().mockResolvedValue(undefined);
    const onCancelEdit = vi.fn();

    render(<TaskForm todo={makeTodo()} onCreate={vi.fn()} onUpdate={onUpdate} onCancelEdit={onCancelEdit} />);

    fireEvent.change(screen.getByPlaceholderText('Task title'), { target: { value: 'Updated title' } });
    fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'done' } });

    fireEvent.click(screen.getByRole('button', { name: 'Update task' }));

    await waitFor(() => {
      expect(onUpdate).toHaveBeenCalledWith(7, {
        title: 'Updated title',
        body: 'Existing body',
        status: 'done',
        priority: 2,
        due_date: '2026-07-01',
      });
    });

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancelEdit).toHaveBeenCalled();
  });
});
