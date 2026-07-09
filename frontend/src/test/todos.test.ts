import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createTodo, deleteTodo, getTodoById, getTodos, updateTodo } from '../api/todos';
import { ApiError, type Todo } from '../types/todo';

const mockFetch = vi.fn();

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
    created_at: '2026-06-20T10:00:00Z',
    updated_at: '2026-06-20T10:00:00Z',
    ...overrides,
  };
}

describe('todos API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('getTodos returns todos on success', async () => {
    const todos: Todo[] = [makeTodo()];

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(todos),
    });

    await expect(getTodos('test-token')).resolves.toEqual(todos);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos', {
      headers: { Authorization: 'Bearer test-token' },
      signal: undefined,
    });
  });

  it('getTodos throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('boom'),
    });

    await expect(getTodos('test-token')).rejects.toEqual(new ApiError(500, 'boom'));
  });

  it('createTodo sends payload, token and returns todo', async () => {
    const created = makeTodo({ id: 2, title: 'Created' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(created),
    });

    await expect(createTodo('test-token', { title: 'Created' })).resolves.toEqual(created);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos', {
      method: 'POST',
      headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
      body: JSON.stringify({ body: '', priority: 0, title: 'Created' }),
    });
  });

  it('getTodoById returns todo on success', async () => {
    const todo = makeTodo({ id: 3, status: 'done' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(todo),
    });

    await expect(getTodoById('test-token', 3)).resolves.toEqual(todo);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/3', {
      headers: { Authorization: 'Bearer test-token' },
    });
  });

  it('updateTodo sends payload, token and returns updated todo', async () => {
    const updated = makeTodo({ id: 4, status: 'done', priority: 3 });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(updated),
    });

    await expect(updateTodo('test-token', 4, { status: 'done' })).resolves.toEqual(updated);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/4', {
      method: 'PUT',
      headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: 'done' }),
    });
  });

  it('deleteTodo returns void for 204 and does not parse json', async () => {
    const jsonSpy = vi.fn();

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: jsonSpy,
    });

    await expect(deleteTodo('test-token', 5)).resolves.toBeUndefined();
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/5', {
      method: 'DELETE',
      headers: { Authorization: 'Bearer test-token' },
    });
    expect(jsonSpy).not.toHaveBeenCalled();
  });

  it('deleteTodo throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('delete failed'),
    });

    await expect(deleteTodo('test-token', 5)).rejects.toEqual(new ApiError(500, 'delete failed'));
  });
});
