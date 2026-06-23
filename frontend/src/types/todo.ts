export interface Todo {
  id: number;
  title: string;
  body: string;
  color: string;
  pinned: boolean;
  done: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTodoPayload {
  title: string;
  body?: string;
  color?: string;
}

export interface UpdateTodoPayload {
  title?: string;
  body?: string;
  color?: string;
  pinned?: boolean;
  done?: boolean;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}
