import { useEffect, useState } from 'react';
import { createTodo, deleteTodo, getTodos, updateTodo } from '../api/todos';
import type { Todo } from '../types/todo';

interface UseTodosReturn {
  todos: Todo[];
  loading: boolean;
  error: string | null;
  addTodo: (title: string, body?: string, color?: string) => Promise<void>;
  editTodo: (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'color' | 'pinned' | 'done'>>,
  ) => Promise<void>;
  toggleTodo: (id: number) => Promise<void>;
  togglePin: (id: number) => Promise<void>;
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

  const addTodo = async (title: string, body = '', color = 'default') => {
    try {
      setError(null);
      const created = await createTodo({ title, body, color });
      setTodos((current) => [...current, created]);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const editTodo = async (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'color' | 'pinned' | 'done'>>,
  ) => {
    const currentTodo = todos.find((todo) => todo.id === id);
    if (!currentTodo) return;

    try {
      setError(null);
      const merged = { ...currentTodo, ...changes };
      const updated = await updateTodo(id, {
        title: merged.title,
        body: merged.body,
        color: merged.color,
        pinned: merged.pinned,
        done: merged.done,
      });
      setTodos((current) =>
        current
          .map((todo) => (todo.id === id ? updated : todo))
          .sort((a, b) => Number(b.pinned) - Number(a.pinned)),
      );
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const toggleTodo = async (id: number) => {
    const currentTodo = todos.find((todo) => todo.id === id);
    if (!currentTodo) return;
    await editTodo(id, { done: !currentTodo.done });
  };

  const togglePin = async (id: number) => {
    const currentTodo = todos.find((todo) => todo.id === id);
    if (!currentTodo) return;
    await editTodo(id, { pinned: !currentTodo.pinned });
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

  return { todos, loading, error, addTodo, editTodo, toggleTodo, togglePin, removeTodo };
}
