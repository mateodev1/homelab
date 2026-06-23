import type { Todo } from '../types/todo';
import { NoteCard } from './NoteCard';

interface NoteGridProps {
  todos: Todo[];
  loading: boolean;
  error: string | null;
  onEdit: (id: number, changes: Partial<Pick<Todo, 'title' | 'body' | 'color' | 'pinned' | 'done'>>) => Promise<void>;
  onDelete: (id: number) => void;
  onTogglePin: (id: number) => void;
}

export function NoteGrid({ todos, loading, error, onEdit, onDelete, onTogglePin }: NoteGridProps) {
  if (loading) {
    return (
      <div className="note-grid__status">
        <div className="spinner" aria-label="Cargando notas" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="note-grid__status note-grid__status--error">
        <p>Error al cargar notas: {error}</p>
      </div>
    );
  }

  const pinned = todos.filter((t) => t.pinned);
  const others = todos.filter((t) => !t.pinned);

  if (todos.length === 0) {
    return (
      <div className="note-grid__empty">
        <div className="note-grid__empty-icon">💡</div>
        <p>Las notas que agregues aparecerán aquí</p>
      </div>
    );
  }

  return (
    <div className="note-grid">
      {pinned.length > 0 && (
        <section className="note-grid__section">
          <h2 className="note-grid__section-label">FIJADAS</h2>
          <div className="note-grid__masonry">
            {pinned.map((todo) => (
              <NoteCard
                key={todo.id}
                todo={todo}
                onEdit={onEdit}
                onDelete={onDelete}
                onTogglePin={onTogglePin}
              />
            ))}
          </div>
        </section>
      )}

      {others.length > 0 && (
        <section className="note-grid__section">
          {pinned.length > 0 && <h2 className="note-grid__section-label">OTRAS</h2>}
          <div className="note-grid__masonry">
            {others.map((todo) => (
              <NoteCard
                key={todo.id}
                todo={todo}
                onEdit={onEdit}
                onDelete={onDelete}
                onTogglePin={onTogglePin}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
