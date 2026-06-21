import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, type Todo } from '../types/todo';
import { createTodo, deleteTodo, getTodoById, getTodos, updateTodo } from './todos';

const mockFetch = vi.fn();

describe('todos API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('getTodos returns todos on success', async () => {
    const todos: Todo[] = [
      {
        id: 1,
        title: 'First todo',
        done: false,
        created_at: '2026-06-20T10:00:00Z',
        updated_at: '2026-06-20T10:00:00Z',
      },
    ];

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
    const created: Todo = {
      id: 2,
      title: 'Created',
      done: false,
      created_at: '2026-06-20T10:10:00Z',
      updated_at: '2026-06-20T10:10:00Z',
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(created),
    });

    await expect(createTodo({ title: 'Created' })).resolves.toEqual(created);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: 'Created' }),
    });
  });

  it('createTodo throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      text: vi.fn().mockResolvedValueOnce('invalid payload'),
    });

    await expect(createTodo({ title: '' })).rejects.toEqual(new ApiError(400, 'invalid payload'));
  });

  it('getTodoById returns todo on success', async () => {
    const todo: Todo = {
      id: 3,
      title: 'Read me',
      done: true,
      created_at: '2026-06-20T10:20:00Z',
      updated_at: '2026-06-20T10:21:00Z',
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(todo),
    });

    await expect(getTodoById(3)).resolves.toEqual(todo);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/3');
  });

  it('getTodoById throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      text: vi.fn().mockResolvedValueOnce('not found'),
    });

    await expect(getTodoById(999)).rejects.toEqual(new ApiError(404, 'not found'));
  });

  it('updateTodo sends payload and returns updated todo', async () => {
    const updated: Todo = {
      id: 4,
      title: 'Updated',
      done: true,
      created_at: '2026-06-20T10:30:00Z',
      updated_at: '2026-06-20T10:40:00Z',
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(updated),
    });

    await expect(updateTodo(4, { done: true })).resolves.toEqual(updated);
    expect(mockFetch).toHaveBeenCalledWith('/api/todos/4', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ done: true }),
    });
  });

  it('updateTodo throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 409,
      text: vi.fn().mockResolvedValueOnce('conflict'),
    });

    await expect(updateTodo(4, { done: true })).rejects.toEqual(new ApiError(409, 'conflict'));
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
