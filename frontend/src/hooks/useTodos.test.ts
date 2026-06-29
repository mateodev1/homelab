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

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
    created_at: '2026-06-21T00:00:00Z',
    updated_at: '2026-06-21T00:00:00Z',
    ...overrides,
  };
}

describe('useTodos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads todos on mount', async () => {
    const todos: Todo[] = [makeTodo({ id: 1, title: 'Write tests first' })];

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
    const created = makeTodo({
      id: 2,
      title: 'New task',
      priority: 2,
      due_date: '2026-07-01',
      created_at: '2026-06-21T00:10:00Z',
      updated_at: '2026-06-21T00:10:00Z',
    });

    mockedGetTodos.mockResolvedValueOnce([]);
    mockedCreateTodo.mockResolvedValueOnce(created);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.addTodo('New task', '', 2, '2026-07-01');
    });

    expect(mockedCreateTodo).toHaveBeenCalledWith({
      title: 'New task',
      body: '',
      priority: 2,
      due_date: '2026-07-01',
    });
    expect(result.current.todos).toEqual([created]);
  });

  it('editTodo updates selected todo', async () => {
    const baseTodo = makeTodo({ id: 3, title: 'Edit me' });
    const updatedTodo = makeTodo({
      ...baseTodo,
      status: 'in_progress',
      priority: 3,
      updated_at: '2026-06-21T00:30:00Z',
    });

    mockedGetTodos.mockResolvedValueOnce([baseTodo]);
    mockedUpdateTodo.mockResolvedValueOnce(updatedTodo);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.editTodo(3, { status: 'in_progress', priority: 3 });
    });

    expect(mockedUpdateTodo).toHaveBeenCalledWith(3, {
      title: 'Edit me',
      body: '',
      status: 'in_progress',
      priority: 3,
      due_date: null,
    });
    expect(result.current.todos).toEqual([updatedTodo]);
  });

  it('removeTodo deletes todo from state', async () => {
    const existing: Todo = makeTodo({ id: 4, title: 'Delete me' });

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

  it('groups todos by status in fixed order and keeps empty groups', async () => {
    mockedGetTodos.mockResolvedValueOnce([
      makeTodo({
        id: 1,
        title: 'Todo high newer',
        status: 'todo',
        priority: 3,
        created_at: '2026-06-22T00:00:00Z',
      }),
      makeTodo({
        id: 2,
        title: 'Todo high older',
        status: 'todo',
        priority: 3,
        created_at: '2026-06-21T00:00:00Z',
      }),
      makeTodo({
        id: 3,
        title: 'In progress',
        status: 'in_progress',
        priority: 1,
      }),
      makeTodo({
        id: 4,
        title: 'Done',
        status: 'done',
        priority: 0,
      }),
    ]);

    const { result } = renderHook(() => useTodos());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    const { groupedTodos } = result.current;

    expect(Object.keys(groupedTodos)).toEqual(['todo', 'in_progress', 'done', 'cancelled']);
    expect(groupedTodos.todo.map((t) => t.title)).toEqual(['Todo high newer', 'Todo high older']);
    expect(groupedTodos.in_progress).toHaveLength(1);
    expect(groupedTodos.done).toHaveLength(1);
    expect(groupedTodos.cancelled).toHaveLength(0);
  });
});
