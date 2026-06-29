import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';
import { ThemeProvider } from './context/ThemeContext';
import { useTodos } from './hooks/useTodos';

const renderApp = () =>
  render(
    <ThemeProvider>
      <App />
    </ThemeProvider>,
  );

vi.mock('./hooks/useTodos', () => ({
  useTodos: vi.fn(),
}));

vi.mock('@uiw/react-md-editor', () => ({
  default: () => <div>Markdown editor</div>,
}));

const mockedUseTodos = vi.mocked(useTodos);

const mockHookBase = {
  loading: false,
  error: null,
  addTodo: vi.fn(),
  editTodo: vi.fn(),
  removeTodo: vi.fn(),
};

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders TaskForm and TaskList using hook state', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'From hook',
          body: '',
          status: 'todo',
          priority: 1,
          due_date: null,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
      groupedTodos: {
        todo: [
          {
            id: 1,
            title: 'From hook',
            body: '',
            status: 'todo',
            priority: 1,
            due_date: null,
            created_at: '2026-06-21T03:00:00Z',
            updated_at: '2026-06-21T03:00:00Z',
          },
        ],
        in_progress: [],
        done: [],
        cancelled: [],
      },
    });

    renderApp();

    expect(screen.getByRole('searchbox', { name: /search tasks/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('renders loading spinner when loading state is true', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [],
      groupedTodos: { todo: [], in_progress: [], done: [], cancelled: [] },
      loading: true,
    });

    renderApp();

    expect(screen.getByLabelText(/loading tasks/i)).toBeInTheDocument();
  });

  it('search filtering hides non-matching tasks', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'Alpha',
          body: '',
          status: 'todo',
          priority: 0,
          due_date: null,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
        {
          id: 2,
          title: 'Beta',
          body: '',
          status: 'todo',
          priority: 0,
          due_date: null,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
      groupedTodos: {
        todo: [
          {
            id: 1,
            title: 'Alpha',
            body: '',
            status: 'todo',
            priority: 0,
            due_date: null,
            created_at: '2026-06-21T03:00:00Z',
            updated_at: '2026-06-21T03:00:00Z',
          },
          {
            id: 2,
            title: 'Beta',
            body: '',
            status: 'todo',
            priority: 0,
            due_date: null,
            created_at: '2026-06-21T03:00:00Z',
            updated_at: '2026-06-21T03:00:00Z',
          },
        ],
        in_progress: [],
        done: [],
        cancelled: [],
      },
    });

    renderApp();

    fireEvent.change(screen.getByRole('searchbox', { name: /search tasks/i }), {
      target: { value: 'alp' },
    });

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
  });
});
