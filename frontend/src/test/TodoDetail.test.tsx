import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TodoDetail } from '../components/TodoDetail';
import { useTodo } from '../hooks/useTodo';
import type { Todo } from '../types/todo';

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

vi.mock('../hooks/useTodo', () => ({
  useTodo: vi.fn(),
}));

vi.mock('../components/RichTextEditor', () => ({
  RichTextEditor: ({
    value,
    onChange,
    onBlur,
  }: {
    value: string;
    onChange: (value: string) => void;
    onBlur?: () => void;
  }) => (
    <textarea
      aria-label="Task body"
      value={value}
      onChange={(event) => onChange(event.target.value)}
      onBlur={onBlur}
    />
  ),
}));

const mockedUseTodo = vi.mocked(useTodo);
const removeTodo = vi.fn().mockResolvedValue(undefined);

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Detail task',
    body: '# Heading\n\nSome **body** text',
    status: 'todo',
    priority: 1,
    due_date: null,
    kind: 'note',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T00:00:00Z',
    updated_at: '2026-06-21T00:00:00Z',
    ...overrides,
  };
}

const save = vi.fn().mockResolvedValue(undefined);

function mockTodo(todo: Todo | null, overrides: Partial<ReturnType<typeof useTodo>> = {}) {
  mockedUseTodo.mockReturnValue({
    todo,
    loading: false,
    error: null,
    save,
    ...overrides,
  });
}

function renderDetail(id = 1) {
  return render(<TodoDetail id={id} projects={[]} removeTodo={removeTodo} />);
}

describe('TodoDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    save.mockClear();
    save.mockResolvedValue(undefined);
    removeTodo.mockClear();
    removeTodo.mockResolvedValue(undefined);
  });

  it('shows a loading spinner while the task loads', () => {
    mockTodo(null, { loading: true });

    renderDetail();

    expect(screen.getByLabelText('Loading task')).toBeInTheDocument();
  });

  it('shows an error message when loading fails', () => {
    mockTodo(null, { error: 'boom' });

    renderDetail();

    expect(screen.getByText('Failed to load task: boom')).toBeInTheDocument();
  });

  it('renders the title, metadata selects and the body editor', () => {
    mockTodo(makeTodo());

    renderDetail();

    expect(screen.getByLabelText('Task title')).toHaveValue('Detail task');
    expect(screen.getByLabelText('Status')).toHaveValue('todo');
    expect(screen.getByLabelText('Priority')).toHaveValue('1');
    expect(screen.getByLabelText('Task body')).toHaveValue('# Heading\n\nSome **body** text');
  });

  it('shows the issue type selector only when kind is issue', () => {
    mockTodo(makeTodo({ kind: 'issue', issue_type: 'bug' }));

    renderDetail();

    expect(screen.getByLabelText('Issue type')).toHaveValue('bug');
  });

  it('debounces title edits and saves after the user stops typing', () => {
    vi.useFakeTimers();
    mockTodo(makeTodo());

    renderDetail();

    fireEvent.change(screen.getByLabelText('Task title'), { target: { value: 'New title' } });
    expect(save).not.toHaveBeenCalled();

    vi.advanceTimersByTime(500);
    expect(save).toHaveBeenCalledWith({ title: 'New title' });

    vi.useRealTimers();
  });

  it('saves immediately on blur', () => {
    mockTodo(makeTodo());

    renderDetail();

    fireEvent.change(screen.getByLabelText('Task title'), { target: { value: 'Blurred title' } });
    fireEvent.blur(screen.getByLabelText('Task title'));

    expect(save).toHaveBeenCalledWith({ title: 'Blurred title' });
  });

  it('debounces body edits and saves after the user stops typing', () => {
    vi.useFakeTimers();
    mockTodo(makeTodo());

    renderDetail();

    fireEvent.change(screen.getByLabelText('Task body'), { target: { value: 'New body' } });
    expect(save).not.toHaveBeenCalled();

    vi.advanceTimersByTime(500);
    expect(save).toHaveBeenCalledWith({ body: 'New body' });

    vi.useRealTimers();
  });

  it('saves immediately when changing a metadata select', () => {
    mockTodo(makeTodo());

    renderDetail();

    fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'done' } });

    expect(save).toHaveBeenCalledWith({ status: 'done' });
  });

  it('deletes the task and navigates back to the list', async () => {
    mockTodo(makeTodo({ id: 8, title: 'Remove me' }));

    renderDetail(8);

    fireEvent.click(screen.getByRole('button', { name: 'Delete Remove me' }));

    await waitFor(() => {
      expect(removeTodo).toHaveBeenCalledWith(8);
    });
    expect(mockedNavigate).toHaveBeenCalledWith({ to: '/todos' });
  });
});
