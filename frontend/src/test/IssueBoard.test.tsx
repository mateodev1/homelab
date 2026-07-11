import type { DragEndEvent } from '@dnd-kit/core';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { IssueBoard, resolveDragStatusChange } from '../components/IssueBoard';
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
    title: 'Issue',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
    kind: 'issue',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

function makeDragEndEvent(overId: string | null, activeId: number): DragEndEvent {
  return {
    active: {
      id: `issue-${activeId}`,
      data: { current: { id: activeId } },
      rect: { current: { initial: null, translated: null } },
    },
    over: overId ? { id: overId, rect: {} as never, data: { current: undefined } } : null,
    collisions: null,
    delta: { x: 0, y: 0 },
    activatorEvent: new Event('pointerup'),
  } as unknown as DragEndEvent;
}

describe('resolveDragStatusChange', () => {
  const issues = [makeTodo({ id: 1, status: 'todo' }), makeTodo({ id: 2, status: 'done' })];

  it('returns the target status when dropped on a different column', () => {
    expect(resolveDragStatusChange(issues, makeDragEndEvent('in_progress', 1))).toEqual({
      id: 1,
      status: 'in_progress',
    });
  });

  it('returns null when dropped on the same column', () => {
    expect(resolveDragStatusChange(issues, makeDragEndEvent('todo', 1))).toBeNull();
  });

  it('returns null when dropped outside any column', () => {
    expect(resolveDragStatusChange(issues, makeDragEndEvent(null, 1))).toBeNull();
  });

  it('returns null when the dragged issue cannot be found', () => {
    expect(resolveDragStatusChange(issues, makeDragEndEvent('done', 999))).toBeNull();
  });
});

describe('IssueBoard', () => {
  it('renders the 4 status columns with issues placed in the right column', () => {
    render(
      <IssueBoard
        issues={[
          makeTodo({ id: 1, title: 'Todo issue', status: 'todo' }),
          makeTodo({ id: 2, title: 'Doing issue', status: 'in_progress' }),
        ]}
        projects={[]}
        activeId={null}
        onSelectTask={vi.fn()}
        onDeleteTask={vi.fn()}
        onStatusChange={vi.fn()}
      />,
    );

    expect(screen.getByTestId('issue-column-todo')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-in_progress')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-done')).toBeInTheDocument();
    expect(screen.getByTestId('issue-column-cancelled')).toBeInTheDocument();
    expect(screen.getByText('Todo issue')).toBeInTheDocument();
    expect(screen.getByText('Doing issue')).toBeInTheDocument();
  });
});
