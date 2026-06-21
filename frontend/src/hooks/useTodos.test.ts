import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { useTodos } from './useTodos';

vi.mock('../api/todos', () => ({
  getTodos: vi.fn(),
  createTodo: vi.fn(),
  updateTodo: vi.fn(),
  deleteTodo: vi.fn(),
}));

import { createTodo, deleteTodo, getTodos, updateTodo } from '../api/todos';

const mockedGetTodos = vi.mocked(getTodos);
const mockedCreateTodo = vi.mocked(createTodo);
const mockedUpdateTodo = vi.mocked(updateTodo);
const mockedDeleteTodo = vi.mocked(deleteTodo);

describe('useTodos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads todos on mount', async () => {
    const todos: Todo[] = [
      {
        id: 1,
        title: 'Write tests first',
        done: false,
        created_at: '2026-06-21T00:00:00Z',
        updated_at: '2026-06-21T00:00:00Z',
      },
    ];

    mockedGetTodos.mockResolvedValueOnce(todos);

    const { result } = renderHook(() => useTodos());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.todos).toEqual(todos);
    expect(result.current.error).toBeNull();
  });

  it('sets error state when initial load fails', async () => {
    mockedGetTodos.mockRejectedValueOnce(new Error('load failed'));

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe('load failed');
    expect(result.current.todos).toEqual([]);
  });

  it('addTodo appends created todo to state', async () => {
    const initial: Todo[] = [];
    const created: Todo = {
      id: 2,
      title: 'New task',
      done: false,
      created_at: '2026-06-21T00:10:00Z',
      updated_at: '2026-06-21T00:10:00Z',
    };

    mockedGetTodos.mockResolvedValueOnce(initial);
    mockedCreateTodo.mockResolvedValueOnce(created);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.addTodo('New task');
    });

    expect(mockedCreateTodo).toHaveBeenCalledWith({ title: 'New task' });
    expect(result.current.todos).toEqual([created]);
  });

  it('toggleTodo updates done state for the selected todo', async () => {
    const baseTodo: Todo = {
      id: 3,
      title: 'Toggle me',
      done: false,
      created_at: '2026-06-21T00:20:00Z',
      updated_at: '2026-06-21T00:20:00Z',
    };
    const toggledTodo: Todo = {
      ...baseTodo,
      done: true,
      updated_at: '2026-06-21T00:30:00Z',
    };

    mockedGetTodos.mockResolvedValueOnce([baseTodo]);
    mockedUpdateTodo.mockResolvedValueOnce(toggledTodo);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleTodo(3);
    });

    expect(mockedUpdateTodo).toHaveBeenCalledWith(3, { done: true });
    expect(result.current.todos).toEqual([toggledTodo]);
  });

  it('removeTodo deletes todo from state', async () => {
    const existing: Todo = {
      id: 4,
      title: 'Delete me',
      done: false,
      created_at: '2026-06-21T00:40:00Z',
      updated_at: '2026-06-21T00:40:00Z',
    };

    mockedGetTodos.mockResolvedValueOnce([existing]);
    mockedDeleteTodo.mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.removeTodo(4);
    });

    expect(mockedDeleteTodo).toHaveBeenCalledWith(4);
    expect(result.current.todos).toEqual([]);
  });

  it('captures action errors in error state', async () => {
    mockedGetTodos.mockResolvedValueOnce([]);
    mockedCreateTodo.mockRejectedValueOnce(new Error('create failed'));

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.addTodo('Boom');
    });

    expect(result.current.error).toBe('create failed');
  });
});
