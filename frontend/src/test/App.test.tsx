import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from '../App';
import { ThemeProvider } from '../context/ThemeContext';
import { useTodos } from '../hooks/useTodos';
import type { Todo } from '../types/todo';

const renderApp = () =>
  render(
    <ThemeProvider>
      <App />
    </ThemeProvider>,
  );

vi.mock('../hooks/useTodos', () => ({
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

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'From hook',
    body: '',
    status: 'todo',
    priority: 1,
    due_date: null,
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

  it('sidebar view filtering shows only tasks from the selected status', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha', status: 'todo' }),
      makeTodo({ id: 2, title: 'Beta', status: 'done' }),
    ]);

    renderApp();

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /^done/i }));

    expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('opens the create task modal from the sidebar button and creates a task', async () => {
    const addTodo = vi.fn().mockResolvedValue(undefined);
    mockTodos([], { addTodo });

    renderApp();

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

    renderApp();

    fireEvent.click(screen.getByRole('button', { name: /new task/i }));
    await screen.findByPlaceholderText('Task title');

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByPlaceholderText('Task title')).not.toBeInTheDocument();
  });

  it('moves the active row with j/k keyboard shortcuts', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha' }),
      makeTodo({ id: 2, title: 'Beta' }),
      makeTodo({ id: 3, title: 'Gamma' }),
    ]);

    renderApp();

    expect(screen.getByTestId('task-row-1')).toHaveClass('bg-accent');

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('task-row-2')).toHaveClass('bg-accent');
    expect(screen.getByTestId('task-row-1')).not.toHaveClass('bg-accent');

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('task-row-3')).toHaveClass('bg-accent');

    fireEvent.keyDown(document, { key: 'k' });
    expect(screen.getByTestId('task-row-2')).toHaveClass('bg-accent');
  });

  it('opens the active task via Enter and closes it via Escape', async () => {
    mockTodos([makeTodo({ id: 1, title: 'Alpha' }), makeTodo({ id: 2, title: 'Beta' })]);

    renderApp();

    fireEvent.keyDown(document, { key: 'j' });
    fireEvent.keyDown(document, { key: 'Enter' });

    const titleInput = await screen.findByPlaceholderText('Task title');
    expect(titleInput).toHaveValue('Beta');

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByPlaceholderText('Task title')).not.toBeInTheDocument();
  });

  it('resets the active row to the first task when the search query changes', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha' }),
      makeTodo({ id: 2, title: 'Beta' }),
      makeTodo({ id: 3, title: 'Alphabet' }),
    ]);

    renderApp();

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('task-row-2')).toHaveClass('bg-accent');

    fireEvent.change(screen.getByRole('searchbox', { name: /search tasks/i }), {
      target: { value: 'alpha' },
    });

    expect(screen.getByTestId('task-row-1')).toHaveClass('bg-accent');
  });

  it('resets the active row to the first task when the sidebar view changes', () => {
    mockTodos([
      makeTodo({ id: 1, title: 'Alpha', status: 'todo' }),
      makeTodo({ id: 2, title: 'Beta', status: 'todo' }),
      makeTodo({ id: 3, title: 'Gamma', status: 'done' }),
    ]);

    renderApp();

    fireEvent.keyDown(document, { key: 'j' });
    expect(screen.getByTestId('task-row-2')).toHaveClass('bg-accent');

    fireEvent.click(screen.getByRole('button', { name: /^done/i }));

    expect(screen.getByTestId('task-row-3')).toHaveClass('bg-accent');
  });
});
