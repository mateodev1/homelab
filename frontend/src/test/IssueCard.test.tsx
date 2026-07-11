import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { IssueCard } from '../components/IssueCard';
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
    title: 'Issue title',
    body: '',
    status: 'todo',
    priority: 2,
    due_date: '2026-07-03',
    kind: 'issue',
    issue_type: 'bug',
    project_id: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

describe('IssueCard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-01T00:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders title, issue type badge and priority badge', () => {
    render(
      <IssueCard
        todo={makeTodo({ priority: 3 })}
        project={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText('Issue title')).toBeInTheDocument();
    expect(screen.getByText('Bug')).toBeInTheDocument();
    expect(screen.getByText('High')).toBeInTheDocument();
  });

  it('renders due date as relative text and hides it when null', () => {
    const { rerender } = render(
      <IssueCard todo={makeTodo()} project={null} onSelect={vi.fn()} onDelete={vi.fn()} />,
    );

    expect(screen.getByText('in 2 days')).toBeInTheDocument();

    rerender(
      <IssueCard
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
      <IssueCard
        todo={makeTodo()}
        project={null}
        active={false}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByTestId('issue-card-1')).not.toHaveClass('ring-2');

    rerender(
      <IssueCard
        todo={makeTodo()}
        project={null}
        active={true}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByTestId('issue-card-1')).toHaveClass('ring-2');
  });

  it('calls onSelect and onDelete', () => {
    const onSelect = vi.fn();
    const onDelete = vi.fn();

    render(<IssueCard todo={makeTodo()} project={null} onSelect={onSelect} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole('button', { name: 'Issue title' }));
    fireEvent.click(screen.getByRole('link', { name: /open issue title/i }));
    fireEvent.click(screen.getByRole('button', { name: /delete issue title/i }));

    expect(onSelect).toHaveBeenCalledWith(1);
    expect(onDelete).toHaveBeenCalledWith(1);
  });

  it('renders an open link pointing to the detail page', () => {
    render(
      <IssueCard
        todo={makeTodo({ id: 42 })}
        project={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByRole('link', { name: /open issue title/i })).toHaveAttribute(
      'href',
      '/todos/42',
    );
  });
});
