import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import App from './App';
import { useTodos } from './hooks/useTodos';

vi.mock('./hooks/useTodos', () => ({
  useTodos: vi.fn(),
}));

const mockedUseTodos = vi.mocked(useTodos);

const mockHookBase = {
  loading: false,
  error: null,
  addTodo: vi.fn(),
  editTodo: vi.fn(),
  toggleTodo: vi.fn(),
  togglePin: vi.fn(),
  removeTodo: vi.fn(),
};

describe('App', () => {
  it('renders NoteForm and NoteGrid using useTodos state', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'From hook',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
    });

    render(<App />);

    expect(screen.getByRole('searchbox', { name: /buscar notas/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('renders loading spinner when loading state is true', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [],
      loading: true,
    });

    render(<App />);

    expect(screen.getByLabelText(/cargando notas/i)).toBeInTheDocument();
  });
});
