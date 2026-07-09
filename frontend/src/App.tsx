import { useAuth0 } from '@auth0/auth0-react';
import { CheckSquare, Moon, Sun, X } from 'lucide-react';
import { useState } from 'react';
import { LoginButton } from './components/LoginButton';
import { LogoutButton } from './components/LogoutButton';
import { TaskForm } from './components/TaskForm';
import { TaskList } from './components/TaskList';
import { Avatar, AvatarImage } from './components/ui/avatar';
import { Button } from './components/ui/button';
import { Input } from './components/ui/input';
import { useTheme } from './context/ThemeContext';
import { useTodos } from './hooks/useTodos';

function App() {
  const { todos, groupedTodos, loading, error, addTodo, editTodo, removeTodo } = useTodos();
  const [query, setQuery] = useState('');
  const [editingTodoID, setEditingTodoID] = useState<number | null>(null);
  const { theme, toggle } = useTheme();
  const { isAuthenticated, user } = useAuth0();

  const filteredGroupedTodos = {
    todo: groupedTodos.todo.filter(matchesQuery(query)),
    in_progress: groupedTodos.in_progress.filter(matchesQuery(query)),
    done: groupedTodos.done.filter(matchesQuery(query)),
    cancelled: groupedTodos.cancelled.filter(matchesQuery(query)),
  };

  const editingTodo =
    editingTodoID == null ? null : (todos.find((todo) => todo.id === editingTodoID) ?? null);

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-10 flex h-16 items-center gap-4 border-b border-border bg-background px-4">
        <div className="flex shrink-0 items-center gap-2">
          <CheckSquare aria-hidden="true" className="size-5 text-foreground" />
          <span className="text-base font-medium tracking-tight text-foreground">Tasks</span>
        </div>

        <div className="relative flex-1 max-w-xl">
          <Input
            type="search"
            placeholder="Search tasks"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            aria-label="Search tasks"
            className="pl-3 pr-8"
          />
          {query && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setQuery('')}
              aria-label="Clear search"
              className="absolute right-0.5 top-0.5 h-8 w-8"
            >
              <X aria-hidden="true" className="size-4" />
            </Button>
          )}
        </div>

        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={toggle}
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? <Sun aria-hidden="true" /> : <Moon aria-hidden="true" />}
        </Button>

        {isAuthenticated ? (
          <>
            {user?.picture && (
              <Avatar>
                <AvatarImage src={user.picture} alt={user.name ?? 'User'} />
              </Avatar>
            )}
            <LogoutButton />
          </>
        ) : (
          <LoginButton />
        )}
      </header>

      <main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-6">
        <TaskForm
          todo={editingTodo}
          onCreate={addTodo}
          onUpdate={async (id, changes) => {
            await editTodo(id, changes);
            setEditingTodoID(null);
          }}
          onCancelEdit={() => setEditingTodoID(null)}
        />

        <TaskList
          groupedTodos={filteredGroupedTodos}
          loading={loading}
          error={error}
          onSelectTask={(id) => setEditingTodoID(id)}
          onDeleteTask={removeTodo}
        />
      </main>
    </div>
  );
}

function matchesQuery(query: string) {
  const normalized = query.trim().toLowerCase();

  return (todo: { title: string; body: string }) => {
    if (!normalized) {
      return true;
    }

    return (
      todo.title.toLowerCase().includes(normalized) || todo.body.toLowerCase().includes(normalized)
    );
  };
}

export default App;
