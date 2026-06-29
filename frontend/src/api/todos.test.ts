import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, type Todo } from '../types/todo';
import { createTodo, deleteTodo, getTodoById, getTodos, updateTodo } from './todos';

const mockFetch = vi.fn();

function makeTodo(overrides: Partial<Todo> = {}): Todo {
  return {
    id: 1,
    title: 'Task',
    body: '',
    status: 'todo',
    priority: 0,
    due_date: null,
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

    await expect(getTodos()).resolves.toEqual(todos);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos');
  });

  it('getTodos throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('boom'),
    });

    await expect(getTodos()).rejects.toEqual(new ApiError(500, 'boom'));
  });

  it('createTodo sends payload and returns todo', async () => {
    const created = makeTodo({ id: 2, title: 'Created' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(created),
    });

    await expect(createTodo({ title: 'Created' })).resolves.toEqual(created);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ body: '', priority: 0, title: 'Created' }),
    });
  });

  it('getTodoById returns todo on success', async () => {
    const todo = makeTodo({ id: 3, status: 'done' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(todo),
    });

    await expect(getTodoById(3)).resolves.toEqual(todo);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/3');
  });

  it('updateTodo sends payload and returns updated todo', async () => {
    const updated = makeTodo({ id: 4, status: 'done', priority: 3 });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(updated),
    });

    await expect(updateTodo(4, { status: 'done' })).resolves.toEqual(updated);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/4', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
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

    await expect(deleteTodo(5)).resolves.toBeUndefined();
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/5', { method: 'DELETE' });
    expect(jsonSpy).not.toHaveBeenCalled();
  });

  it('deleteTodo throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('delete failed'),
    });

    await expect(deleteTodo(5)).rejects.toEqual(new ApiError(500, 'delete failed'));
  });
});
