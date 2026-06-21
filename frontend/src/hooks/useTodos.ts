import { useEffect, useState } from 'react';
import { createTodo, deleteTodo, getTodos, updateTodo } from '../api/todos';
import type { Todo } from '../types/todo';

interface UseTodosReturn {
  todos: Todo[];
  loading: boolean;
  error: string | null;
  addTodo: (title: string) => Promise<void>;
  toggleTodo: (id: number) => Promise<void>;
  removeTodo: (id: number) => Promise<void>;
}

function toMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unknown error';
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

  const addTodo = async (title: string) => {
    try {
      setError(null);
      const created = await createTodo({ title });
      setTodos((current) => [...current, created]);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const toggleTodo = async (id: number) => {
    const currentTodo = todos.find((todo) => todo.id === id);

    if (!currentTodo) {
      return;
    }

    try {
      setError(null);
      const updated = await updateTodo(id, { done: !currentTodo.done });
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

  return { todos, loading, error, addTodo, toggleTodo, removeTodo };
}
