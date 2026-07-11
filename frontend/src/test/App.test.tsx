import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from '../App';
import { ThemeProvider } from '../context/ThemeContext';
import { TodosBoardProvider } from '../context/TodosBoardContext';
import { useProjects } from '../hooks/useProjects';
import { useTodos } from '../hooks/useTodos';
import type { Todo } from '../types/todo';

const renderApp = () =>
  render(
    <ThemeProvider>
      <TodosBoardProvider>
        <App />
      </TodosBoardProvider>
    </ThemeProvider>,
  );

vi.mock('../hooks/useTodos', () => ({
  useTodos: vi.fn(),
}));

vi.mock('../hooks/useProjects', () => ({
  useProjects: vi.fn(),
}));

vi.mock('@uiw/react-md-editor', () => ({
  default: () => <div>Markdown editor</div>,
}));

const mockedNavigate = vi.fn();

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
  useNavigate: () => mockedNavigate,
}));

const mockedUseTodos = vi.mocked(useTodos);
const mockedUseProjects = vi.mocked(useProjects);

const mockHookBase = {
  loading: false,
  error: null,
  addTodo: vi.fn(),
  editTodo: vi.fn(),
  removeTodo: vi.fn(),
};

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'From hook',
    body: '',
    status: 'todo',
    priority: 1,
    due_date: null,
    kind: 'note',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T03:00:00Z',
    updated_at: '2026-06-21T03:00:00Z',
    ...overrides,
  };
}

function groupByStatus(todos: Todo[]) {
  return {
    todo: todos.filter((t) => t.status === 'todo'),
    in_progress: todos.filter((t) => t.status === 'in_progress'),
    done: todos.filter((t) => t.status === 'done'),
    cancelled: todos.filter((t) => t.status === 'cancelled'),
  };
}

function mockTodos(todos: Todo[], overrides: Partial<typeof mockHookBase> = {}) {
  mockedUseTodos.mockReturnValue({
    ...mockHookBase,
    ...overrides,
    todos,
    groupedTodos: groupByStatus(todos),
  });
}

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseProjects.mockReturnValue({
      projects: [],
      loading: false,
      error: null,
      addProject: vi.fn(),
      editProject: vi.fn(),
      removeProject: vi.fn(),
    });
  });

  it('renders the search input and task list using hook state', () => {
    mockTodos([makeTodo()]);

    renderApp();

    expect(screen.getByRole('searchbox', { name: /search tasks/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('renders loading spinner when loading state is true', () => {
    mockTodos([], { loading: true });

    renderApp();

    expect(screen.getByLabelText(/loading tasks/i)).toBeInTheDocument();
  });

  it('search filtering hides non-matching tasks', () => {
    mockTodos([makeTodo({ id: 1, title: 'Alpha' }), makeTodo({ id: 2, title: 'Beta' })]);

    renderApp();

    fireEvent.change(screen.getByRole('searchbox', { name: /search tasks/i }), {
      target: { value: 'alp' },
    });

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
  });

  it('moves the active row with j/k keyboard shortcuts', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha' }),
      makeTodo({ id: 2, title: 'Beta' }),
      makeTodo({ id: 3, title: 'Gamma' }),
    ]);

    renderApp();

    expect(screen.getByTestId('note-card-1')).toHaveClass('ring-2');

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('note-card-2')).toHaveClass('ring-2');
    expect(screen.getByTestId('note-card-1')).not.toHaveClass('ring-2');

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('note-card-3')).toHaveClass('ring-2');

    fireEvent.keyDown(document, { key: 'k' });
    expect(screen.getByTestId('note-card-2')).toHaveClass('ring-2');
  });

  it('resets the active row to the first task when the search query changes', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha' }),
      makeTodo({ id: 2, title: 'Beta' }),
      makeTodo({ id: 3, title: 'Alphabet' }),
    ]);

    renderApp();

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('note-card-2')).toHaveClass('ring-2');

    fireEvent.change(screen.getByRole('searchbox', { name: /search tasks/i }), {
      target: { value: 'alpha' },
    });

    expect(screen.getByTestId('note-card-1')).toHaveClass('ring-2');
  });
});
