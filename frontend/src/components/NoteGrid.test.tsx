import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { NoteGrid } from './NoteGrid';

function makeHandlers() {
  return {
    onEdit: vi.fn().mockResolvedValue(undefined),
    onDelete: vi.fn(),
    onTogglePin: vi.fn(),
  };
}

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Default note',
    body: '',
    color: 'default',
    pinned: false,
    done: false,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

describe('NoteGrid', () => {
  it('loading state renders spinner', () => {
    render(<NoteGrid todos={[]} loading={true} error={null} {...makeHandlers()} />);

    expect(screen.getByLabelText('Cargando notas')).toBeInTheDocument();
  });

  it('error state renders error message', () => {
    render(<NoteGrid todos={[]} loading={false} error="server error" {...makeHandlers()} />);

    expect(screen.getByText('Error al cargar notas: server error')).toBeInTheDocument();
  });

  it('empty state renders empty message', () => {
    render(<NoteGrid todos={[]} loading={false} error={null} {...makeHandlers()} />);

    expect(screen.getByText('Las notas que agregues aparecerán aquí')).toBeInTheDocument();
  });

  it('only unpinned: cards render, no section headers', () => {
    const todos = [
      makeTodo({ id: 1, title: 'Note A', pinned: false }),
      makeTodo({ id: 2, title: 'Note B', pinned: false }),
    ];
    render(<NoteGrid todos={todos} loading={false} error={null} {...makeHandlers()} />);

    expect(screen.getByText('Note A')).toBeInTheDocument();
    expect(screen.getByText('Note B')).toBeInTheDocument();
    expect(screen.queryByText('FIJADAS')).not.toBeInTheDocument();
    expect(screen.queryByText('OTRAS')).not.toBeInTheDocument();
  });

  it('only pinned: shows FIJADAS, no OTRAS', () => {
    const todos = [makeTodo({ id: 1, title: 'Pinned note', pinned: true })];
    render(<NoteGrid todos={todos} loading={false} error={null} {...makeHandlers()} />);

    expect(screen.getByText('FIJADAS')).toBeInTheDocument();
    expect(screen.queryByText('OTRAS')).not.toBeInTheDocument();
  });

  it('mixed pinned + unpinned: shows both FIJADAS and OTRAS', () => {
    const todos = [
      makeTodo({ id: 1, title: 'Pinned note', pinned: true }),
      makeTodo({ id: 2, title: 'Normal note', pinned: false }),
    ];
    render(<NoteGrid todos={todos} loading={false} error={null} {...makeHandlers()} />);

    expect(screen.getByText('FIJADAS')).toBeInTheDocument();
    expect(screen.getByText('OTRAS')).toBeInTheDocument();
    expect(screen.getByText('Pinned note')).toBeInTheDocument();
    expect(screen.getByText('Normal note')).toBeInTheDocument();
  });
});
