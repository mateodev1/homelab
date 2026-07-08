import { ApiError, type CreateTodoPayload, type Todo, type UpdateTodoPayload } from '../types/todo';

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.json() as Promise<T>;
}

function authorizedHeaders(token: string, extra?: Record<string, string>): Record<string, string> {
  return { Authorization: `Bearer ${token}`, ...extra };
}

export async function getTodos(token: string, signal?: AbortSignal): Promise<Todo[]> {
  const response = await fetch('/api/todos', { headers: authorizedHeaders(token), signal });
  return parseResponse<Todo[]>(response);
}

export async function createTodo(token: string, payload: CreateTodoPayload): Promise<Todo> {
  const response = await fetch('/api/todos', {
    method: 'POST',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({ body: '', priority: 0, ...payload }),
  });

  return parseResponse<Todo>(response);
}

export async function getTodoById(token: string, id: number): Promise<Todo> {
  const response = await fetch(`/api/todos/${id}`, { headers: authorizedHeaders(token) });
  return parseResponse<Todo>(response);
}

export async function updateTodo(
  token: string,
  id: number,
  payload: UpdateTodoPayload,
): Promise<Todo> {
  const response = await fetch(`/api/todos/${id}`, {
    method: 'PUT',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify(payload),
  });

  return parseResponse<Todo>(response);
}

export async function deleteTodo(token: string, id: number): Promise<void> {
  const response = await fetch(`/api/todos/${id}`, {
    method: 'DELETE',
    headers: authorizedHeaders(token),
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
}
