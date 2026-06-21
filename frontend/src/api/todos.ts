import { ApiError, type CreateTodoPayload, type Todo, type UpdateTodoPayload } from '../types/todo';

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.json() as Promise<T>;
}

export async function getTodos(signal?: AbortSignal): Promise<Todo[]> {
  const response = signal ? await fetch('/api/todos', { signal }) : await fetch('/api/todos');
  return parseResponse<Todo[]>(response);
}

export async function createTodo(payload: CreateTodoPayload): Promise<Todo> {
  const response = await fetch('/api/todos', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  return parseResponse<Todo>(response);
}

export async function getTodoById(id: number): Promise<Todo> {
  const response = await fetch(`/api/todos/${id}`);
  return parseResponse<Todo>(response);
}

export async function updateTodo(id: number, payload: UpdateTodoPayload): Promise<Todo> {
  const response = await fetch(`/api/todos/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  return parseResponse<Todo>(response);
}

export async function deleteTodo(id: number): Promise<void> {
  const response = await fetch(`/api/todos/${id}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
}
