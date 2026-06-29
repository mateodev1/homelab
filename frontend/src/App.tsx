import { useAuth0 } from '@auth0/auth0-react';
import { useState } from 'react';
import { LoginButton } from './components/LoginButton';
import { LogoutButton } from './components/LogoutButton';
import { TaskForm } from './components/TaskForm';
import { TaskList } from './components/TaskList';
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
    <div className="app">
      <header className="app-header">
        <div className="app-header__logo">
          <span className="app-header__logo-icon">✅</span>
          <span className="app-header__logo-text">Tasks</span>
        </div>
        <div className="app-header__search">
          <span className="app-header__search-icon">🔍</span>
          <input
            className="app-header__search-input"
            type="search"
            placeholder="Search tasks"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            aria-label="Search tasks"
          />
          {query && (
            <button
              type="button"
              className="app-header__search-clear"
              onClick={() => setQuery('')}
              aria-label="Clear search"
            >
              ✕
            </button>
          )}
        </div>
        <button
          type="button"
          className="app-header__theme-toggle"
          onClick={toggle}
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? '☀️' : '🌙'}
        </button>
        {isAuthenticated ? (
          <>
            {user?.picture && (
              <img
                src={user.picture}
                alt={user.name ?? 'User'}
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: '50%',
                  objectFit: 'cover',
                  border: '1px solid var(--color-border)',
                }}
              />
            )}
            <LogoutButton />
          </>
        ) : (
          <LoginButton />
        )}
      </header>

      <main className="app-main app-main--tasks">
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
