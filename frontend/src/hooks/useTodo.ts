import { useAuth0 } from '@auth0/auth0-react';
import { useEffect, useState } from 'react';
import { getTodoById, updateTodo as updateTodoRequest } from '../api/todos';
import type { Todo } from '../types/todo';

interface UseTodoReturn {
  todo: Todo | null;
  loading: boolean;
  error: string | null;
  save: (changes: Partial<Pick<Todo, keyof Todo>>) => Promise<void>;
}

function toMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unknown error';
}

export function useTodo(id: number): UseTodoReturn {
  const { getAccessTokenSilently } = useAuth0();
  const [todo, setTodo] = useState<Todo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);

    const loadTodo = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getTodoById(token, id);
        if (!controller.signal.aborted) {
          setTodo(data);
        }
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

    void loadTodo();

    return () => {
      controller.abort();
    };
  }, [id, getAccessTokenSilently]);

  const save = async (changes: Partial<Todo>) => {
    if (!todo) return;

    try {
      setError(null);
      const merged = { ...todo, ...changes };
      const token = await getAccessTokenSilently();
      const updated = await updateTodoRequest(token, id, {
        title: merged.title,
        body: merged.body,
        status: merged.status,
        priority: merged.priority,
        due_date: merged.due_date,
        kind: merged.kind,
        issue_type: merged.issue_type,
        project_id: merged.project_id,
      });
      setTodo(updated);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  return { todo, loading, error, save };
}
