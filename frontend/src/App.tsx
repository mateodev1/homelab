import { useState } from 'react';
import { NoteForm } from './components/NoteForm';
import { NoteGrid } from './components/NoteGrid';
import { useTheme } from './context/ThemeContext';
import { useTodos } from './hooks/useTodos';

function App() {
  const { todos, loading, error, addTodo, editTodo, removeTodo, togglePin } = useTodos();
  const [query, setQuery] = useState('');
  const { theme, toggle } = useTheme();

  const filtered = query.trim()
    ? todos.filter(
        (t) =>
          t.title.toLowerCase().includes(query.toLowerCase()) ||
          t.body.toLowerCase().includes(query.toLowerCase()),
      )
    : todos;

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header__logo">
          <span className="app-header__logo-icon">💡</span>
          <span className="app-header__logo-text">Keep</span>
        </div>
        <div className="app-header__search">
          <span className="app-header__search-icon">🔍</span>
          <input
            className="app-header__search-input"
            type="search"
            placeholder="Buscar"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            aria-label="Buscar notas"
          />
          {query && (
            <button
              type="button"
              className="app-header__search-clear"
              onClick={() => setQuery('')}
              aria-label="Limpiar búsqueda"
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
      </header>

      <main className="app-main">
        <NoteForm onAdd={addTodo} />
        <NoteGrid
          todos={filtered}
          loading={loading}
          error={error}
          onEdit={editTodo}
          onDelete={removeTodo}
          onTogglePin={togglePin}
        />
      </main>
    </div>
  );
}

export default App;
