import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { NoteCard } from '../components/NoteCard';
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
    title: 'Note title',
    body: '# First paragraph\n\nSecond paragraph',
    status: 'todo',
    priority: 2,
    due_date: '2026-07-03',
    kind: 'note',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

describe('NoteCard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-01T00:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders title, a plain-text body preview and priority badge', () => {
    render(
      <NoteCard
        todo={makeTodo({ priority: 3 })}
        project={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText('Note title')).toBeInTheDocument();
    expect(screen.getByText(/First paragraph Second paragraph/)).toBeInTheDocument();
    expect(screen.getByText('High')).toBeInTheDocument();
  });

  it('renders due date as relative text and hides it when null', () => {
    const { rerender } = render(
      <NoteCard todo={makeTodo()} project={null} onSelect={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByText('in 2 days')).toBeInTheDocument();

    rerender(
      <NoteCard
        todo={makeTodo({ due_date: null })}
        project={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.queryByText(/in \d+ days/i)).not.toBeInTheDocument();
  });

  it('applies active highlight styling when active', () => {
    const { rerender } = render(
      <NoteCard
        todo={makeTodo()}
        project={null}
        active={false}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByTestId('note-card-1')).not.toHaveClass('ring-2');

    rerender(
      <NoteCard
        todo={makeTodo()}
        project={null}
        active={true}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByTestId('note-card-1')).toHaveClass('ring-2');
  });

  it('calls onSelect and onDelete', () => {
    const onSelect = vi.fn();
    const onDelete = vi.fn();

    render(<NoteCard todo={makeTodo()} project={null} onSelect={onSelect} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole('button', { name: 'Note title' }));
    fireEvent.click(screen.getByRole('link', { name: /open note title/i }));
    fireEvent.click(screen.getByRole('button', { name: /delete note title/i }));

    expect(onSelect).toHaveBeenCalledWith(1);
    expect(onDelete).toHaveBeenCalledWith(1);
  });

  it('renders an open link pointing to the detail page', () => {
    render(
      <NoteCard todo={makeTodo({ id: 42 })} project={null} onSelect={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByRole('link', { name: /open note title/i })).toHaveAttribute(
      'href',
      '/todos/42',
    );
  });
});
