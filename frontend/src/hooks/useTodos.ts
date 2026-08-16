import { useAuth0 } from '@auth0/auth0-react';
import { useEffect, useMemo, useState } from 'react';
import { createTodo, deleteTodo, getTodos, updateTodo } from '../api/todos';
import type { IssueType, Todo, TodoKind, TodoStatus } from '../types/todo';

interface GroupedTodos {
  todo: Todo[];
  in_progress: Todo[];
  done: Todo[];
  cancelled: Todo[];
}

interface UseTodosReturn {
  todos: Todo[];
  groupedTodos: GroupedTodos;
  loading: boolean;
  error: string | null;
  addTodo: (
    title: string,
    body?: string,
    priority?: 0 | 1 | 2 | 3,
    dueDate?: string | null,
    kind?: TodoKind,
    issueType?: IssueType | null,
    projectId?: number | null,
  ) => Promise<void>;
  editTodo: (
    id: number,
    changes: Partial<
      Pick<
        Todo,
        'title' | 'body' | 'status' | 'priority' | 'due_date' | 'kind' | 'issue_type' | 'project_id'
      >
    >,
  ) => Promise<void>;
  removeTodo: (id: number) => Promise<void>;
}

const STATUS_ORDER: TodoStatus[] = ['todo', 'in_progress', 'done', 'cancelled'];
export const TODO_REFRESH_INTERVAL_MS = 30_000;

function toMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unknown error';
}

function sortGroup(a: Todo, b: Todo): number {
  if (a.priority !== b.priority) {
    return b.priority - a.priority;
  }

  return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
}

function emptyGroups(): GroupedTodos {
  return {
    todo: [],
    in_progress: [],
    done: [],
    cancelled: [],
  };
}

export function useTodos(): UseTodosReturn {
  const { getAccessTokenSilently } = useAuth0();
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    let isInitialLoad = true;
    let requestInFlight = false;

    const loadTodos = async () => {
      if (requestInFlight) return;
      requestInFlight = true;

      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getTodos(token, controller.signal);
        setTodos(data);
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(toMessage(err));
      } finally {
        requestInFlight = false;
        if (isInitialLoad && !controller.signal.aborted) {
          setLoading(false);
        }
        isInitialLoad = false;
      }
    };

    void loadTodos();
    const refreshInterval = window.setInterval(() => void loadTodos(), TODO_REFRESH_INTERVAL_MS);

    return () => {
      window.clearInterval(refreshInterval);
      controller.abort();
    };
  }, [getAccessTokenSilently]);

  const addTodo = async (
    title: string,
    body = '',
    priority: 0 | 1 | 2 | 3 = 0,
    dueDate: string | null = null,
    kind: TodoKind = 'note',
    issueType: IssueType | null = null,
    projectId: number | null = null,
  ) => {
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      const created = await createTodo(token, {
        title,
        body,
        priority,
        due_date: dueDate,
        kind,
        issue_type: issueType,
        project_id: projectId,
      });
      setTodos((current) => [...current, created]);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const editTodo = async (
    id: number,
    changes: Partial<
      Pick<
        Todo,
        'title' | 'body' | 'status' | 'priority' | 'due_date' | 'kind' | 'issue_type' | 'project_id'
      >
    >,
  ) => {
    const currentTodo = todos.find((todo) => todo.id === id);
    if (!currentTodo) return;

    try {
      setError(null);
      const merged = { ...currentTodo, ...changes };
      const token = await getAccessTokenSilently();
      const updated = await updateTodo(token, id, {
        title: merged.title,
        body: merged.body,
        status: merged.status,
        priority: merged.priority,
        due_date: merged.due_date,
        kind: merged.kind,
        issue_type: merged.issue_type,
        project_id: merged.project_id,
      });
      setTodos((current) => current.map((todo) => (todo.id === id ? updated : todo)));
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const removeTodo = async (id: number) => {
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      await deleteTodo(token, id);
      setTodos((current) => current.filter((todo) => todo.id !== id));
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const groupedTodos = useMemo(() => {
    const groups = emptyGroups();

    for (const todo of todos) {
      groups[todo.status].push(todo);
    }

    for (const status of STATUS_ORDER) {
      groups[status].sort(sortGroup);
    }

    return groups;
  }, [todos]);

  return { todos, groupedTodos, loading, error, addTodo, editTodo, removeTodo };
}
