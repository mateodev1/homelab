import { useEffect, useRef, useState } from 'react';
import type { Todo } from '../types/todo';
import { COLOR_MAP } from './colors';

interface NoteCardProps {
  todo: Todo;
  onEdit: (
    id: number,
    changes: Partial<Pick<Todo, 'title' | 'body' | 'color' | 'pinned' | 'done'>>,
  ) => Promise<void>;
  onDelete: (id: number) => void;
  onTogglePin: (id: number) => void;
}

export function NoteCard({ todo, onEdit, onDelete, onTogglePin }: NoteCardProps) {
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingBody, setEditingBody] = useState(false);
  const [titleValue, setTitleValue] = useState(todo.title);
  const [bodyValue, setBodyValue] = useState(todo.body);
  const [showColorPicker, setShowColorPicker] = useState(false);
  const titleRef = useRef<HTMLTextAreaElement>(null);
  const bodyRef = useRef<HTMLTextAreaElement>(null);
  const cardRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setTitleValue(todo.title);
    setBodyValue(todo.body);
  }, [todo.title, todo.body]);

  useEffect(() => {
    if (editingTitle && titleRef.current) {
      titleRef.current.focus();
      titleRef.current.selectionStart = titleRef.current.value.length;
    }
  }, [editingTitle]);

  useEffect(() => {
    if (editingBody && bodyRef.current) {
      bodyRef.current.focus();
    }
  }, [editingBody]);

  const commitTitle = () => {
    setEditingTitle(false);
    const trimmed = titleValue.trim();
    if (trimmed && trimmed !== todo.title) {
      void onEdit(todo.id, { title: trimmed });
    } else {
      setTitleValue(todo.title);
    }
  };

  const commitBody = () => {
    setEditingBody(false);
    if (bodyValue !== todo.body) {
      void onEdit(todo.id, { body: bodyValue });
    }
  };

  const handleColorSelect = (color: string) => {
    setShowColorPicker(false);
    void onEdit(todo.id, { color });
  };

  const bg: string = COLOR_MAP[todo.color] ?? COLOR_MAP.default;

  return (
    <div
      ref={cardRef}
      className={`note-card${todo.done ? ' note-card--done' : ''}`}
      style={{ backgroundColor: bg }}
    >
      <div className="note-card__actions-top">
        <button
          type="button"
          className={`note-card__pin${todo.pinned ? ' note-card__pin--active' : ''}`}
          onClick={() => onTogglePin(todo.id)}
          title={todo.pinned ? 'Desfijar' : 'Fijar nota'}
          aria-label={todo.pinned ? 'Desfijar nota' : 'Fijar nota'}
        >
          📌
        </button>
      </div>

      {editingTitle ? (
        <textarea
          ref={titleRef}
          className="note-card__title-input"
          value={titleValue}
          onChange={(e) => setTitleValue(e.target.value)}
          onBlur={commitTitle}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault();
              commitTitle();
            }
            if (e.key === 'Escape') {
              setTitleValue(todo.title);
              setEditingTitle(false);
            }
          }}
          rows={1}
          style={{ backgroundColor: bg }}
        />
      ) : (
        <button
          type="button"
          className="note-card__title"
          onClick={() => setEditingTitle(true)}
          aria-label="Editar título"
        >
          {todo.title}
        </button>
      )}

      {editingBody ? (
        <textarea
          ref={bodyRef}
          className="note-card__body-input"
          value={bodyValue}
          onChange={(e) => setBodyValue(e.target.value)}
          onBlur={commitBody}
          onKeyDown={(e) => {
            if (e.key === 'Escape') {
              setBodyValue(todo.body);
              setEditingBody(false);
            }
          }}
          placeholder="Agregar nota..."
          style={{ backgroundColor: bg }}
        />
      ) : (
        <button
          type="button"
          className={`note-card__body${!todo.body ? ' note-card__body--empty' : ''}`}
          onClick={() => setEditingBody(true)}
          aria-label="Editar descripción"
        >
          {todo.body || 'Agregar nota...'}
        </button>
      )}

      <div className="note-card__footer">
        <div className="note-card__footer-actions">
          <button
            type="button"
            className={`note-card__check${todo.done ? ' note-card__check--done' : ''}`}
            onClick={() => void onEdit(todo.id, { done: !todo.done })}
            title={todo.done ? 'Marcar pendiente' : 'Marcar completada'}
            aria-label={todo.done ? 'Marcar pendiente' : 'Marcar completada'}
          >
            {todo.done ? '✓' : '○'}
          </button>

          <div className="note-card__color-picker-wrapper">
            <button
              type="button"
              className="note-card__color-btn"
              onClick={() => setShowColorPicker((v) => !v)}
              title="Cambiar color"
              aria-label="Cambiar color"
            >
              🎨
            </button>
            {showColorPicker && (
              <div className="color-picker-popup">
                {Object.entries(COLOR_MAP).map(([key, value]) => (
                  <button
                    key={key}
                    type="button"
                    className={`color-swatch${todo.color === key ? ' color-swatch--active' : ''}`}
                    style={{ backgroundColor: value }}
                    onClick={() => handleColorSelect(key)}
                    title={key}
                    aria-label={`Color ${key}`}
                  />
                ))}
              </div>
            )}
          </div>

          <button
            type="button"
            className="note-card__delete"
            onClick={() => onDelete(todo.id)}
            title="Eliminar nota"
            aria-label="Eliminar nota"
          >
            🗑️
          </button>
        </div>
      </div>
    </div>
  );
}
