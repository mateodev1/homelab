import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from '../App';
import { TodosLayout } from '../components/TodosLayout';
import { ThemeProvider } from '../context/ThemeContext';
import { useProjects } from '../hooks/useProjects';
import { useTodos } from '../hooks/useTodos';
import type { Todo } from '../types/todo';

const renderLayout = () =>
  render(
    <ThemeProvider>
      <TodosLayout>
        <App />
      </TodosLayout>
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

describe('TodosLayout', () => {
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

  it('renders the sidebar alongside the board content', () => {
    mockTodos([makeTodo()]);

    renderLayout();

    expect(screen.getByRole('navigation', { name: /task views/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('sidebar view filtering shows only tasks from the selected status', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha', status: 'todo' }),
      makeTodo({ id: 2, title: 'Beta', status: 'done' }),
    ]);

    renderLayout();

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /^done/i }));

    expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('sidebar kind filtering shows only tasks from the selected kind', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'A note', kind: 'note' }),
      makeTodo({ id: 2, title: 'An issue', kind: 'issue', issue_type: 'bug' }),
    ]);

    renderLayout();

    expect(screen.getByText('A note')).toBeInTheDocument();
    expect(screen.getByText('An issue')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /^issues/i }));

    expect(screen.queryByText('A note')).not.toBeInTheDocument();
    expect(screen.getByText('An issue')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /^notes/i }));

    expect(screen.getByText('A note')).toBeInTheDocument();
    expect(screen.queryByText('An issue')).not.toBeInTheDocument();
  });

  it('opens the create task modal from the sidebar button and creates a task', async () => {
    const addTodo = vi.fn().mockResolvedValue(undefined);
    mockTodos([], { addTodo });

    renderLayout();

    expect(screen.queryByPlaceholderText('Task title')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /new task/i }));

    const titleInput = await screen.findByPlaceholderText('Task title');
    fireEvent.change(titleInput, { target: { value: 'New task' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add task' }));

    await waitFor(() => {
      expect(addTodo).toHaveBeenCalled();
    });
  });

  it('closes the modal on Escape', async () => {
    mockTodos([]);

    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /new task/i }));
    await screen.findByPlaceholderText('Task title');

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByPlaceholderText('Task title')).not.toBeInTheDocument();
  });

  it('opens the active task via Enter and closes it via Escape', async () => {
    mockTodos([makeTodo({ id: 1, title: 'Alpha' }), makeTodo({ id: 2, title: 'Beta' })]);

    renderLayout();

    fireEvent.keyDown(document, { key: 'j' });
    fireEvent.keyDown(document, { key: 'Enter' });

    const titleInput = await screen.findByPlaceholderText('Task title');
    expect(titleInput).toHaveValue('Beta');

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByPlaceholderText('Task title')).not.toBeInTheDocument();
  });

  it('resets the active row to the first task when the sidebar view changes', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha', status: 'todo' }),
      makeTodo({ id: 2, title: 'Beta', status: 'todo' }),
      makeTodo({ id: 3, title: 'Gamma', status: 'done' }),
    ]);

    renderLayout();

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('note-card-2')).toHaveClass('ring-2');

    fireEvent.click(screen.getByRole('button', { name: /^done/i }));

    expect(screen.getByTestId('note-card-3')).toHaveClass('ring-2');
  });
});
