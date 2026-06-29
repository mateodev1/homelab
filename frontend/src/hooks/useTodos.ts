import { useEffect, useMemo, useState } from 'react';
import { createTodo, deleteTodo, getTodos, updateTodo } from '../api/todos';
import type { Todo, TodoStatus } from '../types/todo';

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
  ) => Promise<void>;
  editTodo: (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'status' | 'priority' | 'due_date'>>,
  ) => Promise<void>;
  removeTodo: (id: number) => Promise<void>;
}

const STATUS_ORDER: TodoStatus[] = ['todo', 'in_progress', 'done', 'cancelled'];

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
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    const loadTodos = async () => {
      try {
        setError(null);
        const data = await getTodos(controller.signal);
        setTodos(data);
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(toMessage(err));
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };

    void loadTodos();

    return () => {
      controller.abort();
    };
  }, []);

  const addTodo = async (
    title: string,
    body = '',
    priority: 0 | 1 | 2 | 3 = 0,
    dueDate: string | null = null,
  ) => {
    try {
      setError(null);
      const created = await createTodo({ title, body, priority, due_date: dueDate });
      setTodos((current) => [...current, created]);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const editTodo = async (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'status' | 'priority' | 'due_date'>>,
  ) => {
    const currentTodo = todos.find((todo) => todo.id === id);
    if (!currentTodo) return;

    try {
      setError(null);
      const merged = { ...currentTodo, ...changes };
      const updated = await updateTodo(id, {
        title: merged.title,
        body: merged.body,
        status: merged.status,
        priority: merged.priority,
        due_date: merged.due_date,
      });
      setTodos((current) => current.map((todo) => (todo.id === id ? updated : todo)));
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const removeTodo = async (id: number) => {
    try {
      setError(null);
      await deleteTodo(id);
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
