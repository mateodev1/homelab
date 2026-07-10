export type TodoStatus = 'todo' | 'in_progress' | 'done' | 'cancelled';
export type Priority = 0 | 1 | 2 | 3;
export type TodoKind = 'note' | 'issue';
export type IssueType = 'feature' | 'bug' | 'improvement';

export interface Todo {
  id: number;
  title: string;
  body: string;
  status: TodoStatus;
  priority: Priority;
  due_date: string | null;
  kind: TodoKind;
  issue_type: IssueType | null;
  project_id: number | null;
  created_at: string;
  updated_at: string;
}

export interface CreateTodoPayload {
  title: string;
  body?: string;
  priority?: Priority;
  due_date?: string | null;
  kind?: TodoKind;
  issue_type?: IssueType | null;
  project_id?: number | null;
}

export interface UpdateTodoPayload {
  title?: string;
  body?: string;
  status?: TodoStatus;
  priority?: Priority;
  due_date?: string | null;
  kind?: TodoKind;
  issue_type?: IssueType | null;
  project_id?: number | null;
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
