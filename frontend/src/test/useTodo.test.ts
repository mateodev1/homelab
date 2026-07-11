import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useTodo } from '../hooks/useTodo';
import type { Todo } from '../types/todo';

vi.mock('../api/todos', () => ({
  getTodoById: vi.fn(),
  updateTodo: vi.fn(),
}));

const mockedGetAccessTokenSilently = vi.fn();

vi.mock('@auth0/auth0-react', () => ({
  useAuth0: () => ({ getAccessTokenSilently: mockedGetAccessTokenSilently }),
}));

import { getTodoById, updateTodo } from '../api/todos';

const mockedGetTodoById = vi.mocked(getTodoById);
const mockedUpdateTodo = vi.mocked(updateTodo);

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
    kind: 'note',
    issue_type: null,
    project_id: null,
    created_at: '2026-06-21T00:00:00Z',
    updated_at: '2026-06-21T00:00:00Z',
    ...overrides,
  };
}

describe('useTodo', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetAccessTokenSilently.mockResolvedValue('test-token');
  });

  it('loads the todo by id on mount', async () => {
    const todo = makeTodo({ id: 5, title: 'Load me' });
    mockedGetTodoById.mockResolvedValueOnce(todo);

    const { result } = renderHook(() => useTodo(5));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(mockedGetTodoById).toHaveBeenCalledWith('test-token', 5);
    expect(result.current.todo).toEqual(todo);
    expect(result.current.error).toBeNull();
  });

  it('sets error state when the load fails', async () => {
    mockedGetTodoById.mockRejectedValueOnce(new Error('not found'));

    const { result } = renderHook(() => useTodo(9));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe('not found');
    expect(result.current.todo).toBeNull();
  });

  it('save merges changes and persists them via updateTodo', async () => {
    const todo = makeTodo({ id: 3, title: 'Edit me' });
    const updated = makeTodo({ ...todo, status: 'done', priority: 3 });

    mockedGetTodoById.mockResolvedValueOnce(todo);
    mockedUpdateTodo.mockResolvedValueOnce(updated);

    const { result } = renderHook(() => useTodo(3));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.save({ status: 'done', priority: 3 });
    });

    expect(mockedUpdateTodo).toHaveBeenCalledWith('test-token', 3, {
      title: 'Edit me',
      body: '',
      status: 'done',
      priority: 3,
      due_date: null,
      kind: 'note',
      issue_type: null,
      project_id: null,
    });
    expect(result.current.todo).toEqual(updated);
  });

  it('save is a no-op when the todo has not loaded yet', async () => {
    mockedGetTodoById.mockResolvedValueOnce(null as unknown as Todo);

    const { result } = renderHook(() => useTodo(4));

    await act(async () => {
      await result.current.save({ title: 'Ignored' });
    });

    expect(mockedUpdateTodo).not.toHaveBeenCalled();
  });
});
