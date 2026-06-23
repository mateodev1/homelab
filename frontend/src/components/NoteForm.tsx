import { type KeyboardEvent, useEffect, useRef, useState } from 'react';
import { COLOR_MAP } from './colors';

interface NoteFormProps {
  onAdd: (title: string, body?: string, color?: string) => void;
}

export function NoteForm({ onAdd }: NoteFormProps) {
  const [expanded, setExpanded] = useState(false);
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [color, setColor] = useState('default');
  const [showColorPicker, setShowColorPicker] = useState(false);
  const formRef = useRef<HTMLDivElement>(null);
  const titleRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (expanded && titleRef.current) {
      titleRef.current.focus();
    }
  }, [expanded]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (formRef.current && !formRef.current.contains(e.target as Node)) {
        commit();
      }
    };
    if (expanded) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  });

  const commit = () => {
    const trimmed = title.trim();
    if (trimmed) {
      onAdd(trimmed, body, color);
    }
    setTitle('');
    setBody('');
    setColor('default');
    setExpanded(false);
    setShowColorPicker(false);
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      setTitle('');
      setBody('');
      setColor('default');
      setExpanded(false);
      setShowColorPicker(false);
    }
  };

  const bg: string = COLOR_MAP[color] ?? COLOR_MAP.default;

  return (
    <div
      ref={formRef}
      className="note-form"
      style={{ backgroundColor: bg }}
      onKeyDown={handleKeyDown}
    >
      {expanded && (
        <input
          ref={titleRef}
          className="note-form__title"
          type="text"
          placeholder="Título"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          style={{ backgroundColor: bg }}
        />
      )}

      <div className="note-form__body-row">
        {!expanded ? (
          <button
            type="button"
            className="note-form__placeholder"
            onClick={() => setExpanded(true)}
            aria-label="Nueva nota"
          >
            Tomar una nota...
          </button>
        ) : (
          <textarea
            className="note-form__body"
            placeholder="Crear nota..."
            value={body}
            onChange={(e) => setBody(e.target.value)}
            rows={3}
            style={{ backgroundColor: bg }}
          />
        )}
      </div>

      {expanded && (
        <div className="note-form__footer">
          <div className="note-form__color-picker-wrapper">
            <button
              type="button"
              className="note-form__icon-btn"
              onClick={() => setShowColorPicker((v) => !v)}
              title="Cambiar color"
              aria-label="Cambiar color de la nota"
            >
              🎨
            </button>
            {showColorPicker && (
              <div className="color-picker-popup">
                {Object.entries(COLOR_MAP).map(([key, value]) => (
                  <button
                    key={key}
                    type="button"
                    className={`color-swatch${color === key ? ' color-swatch--active' : ''}`}
                    style={{ backgroundColor: value }}
                    onClick={() => {
                      setColor(key);
                      setShowColorPicker(false);
                    }}
                    title={key}
                    aria-label={`Color ${key}`}
                  />
                ))}
              </div>
            )}
          </div>

          <button
            type="button"
            className="note-form__close-btn"
            onClick={commit}
          >
            Cerrar
          </button>
        </div>
      )}
    </div>
  );
}
