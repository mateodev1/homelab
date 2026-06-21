import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import App from './App';
import { useTodos } from './hooks/useTodos';

vi.mock('./hooks/useTodos', () => ({
  useTodos: vi.fn(),
}));

const mockedUseTodos = vi.mocked(useTodos);

describe('App', () => {
  it('renders TodoForm and TodoList using useTodos state', () => {
    mockedUseTodos.mockReturnValue({
      todos: [
        {
          id: 1,
          title: 'From hook',
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
      loading: false,
      error: null,
      addTodo: vi.fn(),
      toggleTodo: vi.fn(),
      removeTodo: vi.fn(),
    });

    render(<App />);

    const heading = screen.getByRole('heading', { name: /todo app/i });
    expect(heading).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: /todo title/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('renders without crashing when loading state is true', () => {
    mockedUseTodos.mockReturnValue({
      todos: [],
      loading: true,
      error: null,
      addTodo: vi.fn(),
      toggleTodo: vi.fn(),
      removeTodo: vi.fn(),
    });

    render(<App />);

    expect(screen.getByRole('heading', { name: /todo app/i })).toBeInTheDocument();
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });
});
