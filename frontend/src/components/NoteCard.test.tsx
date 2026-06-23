import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Todo } from '../types/todo';
import { NoteCard } from './NoteCard';

const baseTodo: Todo = {
  id: 1,
  title: 'Test title',
  body: 'Test body',
  color: 'default',
  pinned: false,
  done: false,
  created_at: '2026-06-21T03:00:00Z',
  updated_at: '2026-06-21T03:00:00Z',
};

function makeHandlers() {
  return {
    onEdit: vi.fn().mockResolvedValue(undefined),
    onDelete: vi.fn(),
    onTogglePin: vi.fn(),
  };
}

describe('NoteCard', () => {
  it('renders title and body', () => {
    render(<NoteCard todo={baseTodo} {...makeHandlers()} />);

    expect(screen.getByText('Test title')).toBeInTheDocument();
    expect(screen.getByText('Test body')).toBeInTheDocument();
  });

  it('pin button calls onTogglePin with todo id', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Fijar nota' }));

    expect(handlers.onTogglePin).toHaveBeenCalledWith(1);
  });

  it('done button calls onEdit with done: true when todo is not done', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Marcar completada' }));

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { done: true });
  });

  it('delete button calls onDelete with todo id', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Eliminar nota' }));

    expect(handlers.onDelete).toHaveBeenCalledWith(1);
  });

  it('clicking title button shows textarea; blur with new value calls onEdit', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar título' }));

    const textarea = screen.getByDisplayValue('Test title');
    expect(textarea).toBeInTheDocument();

    fireEvent.change(textarea, { target: { value: 'New title' } });
    fireEvent.blur(textarea);

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { title: 'New title' });
  });

  it('blur with same value does NOT call onEdit (revert path)', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar título' }));

    const textarea = screen.getByDisplayValue('Test title');
    // value unchanged — blur should not trigger onEdit
    fireEvent.blur(textarea);

    expect(handlers.onEdit).not.toHaveBeenCalled();
  });

  it('Escape on title textarea reverts and exits editing', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar título' }));

    const textarea = screen.getByDisplayValue('Test title');
    fireEvent.change(textarea, { target: { value: 'Discard me' } });
    fireEvent.keyDown(textarea, { key: 'Escape' });

    // Title button should be back visible with original value
    expect(screen.getByRole('button', { name: 'Editar título' })).toBeInTheDocument();
    expect(handlers.onEdit).not.toHaveBeenCalled();
  });

  it('Enter key on title textarea commits and calls onEdit', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar título' }));

    const textarea = screen.getByDisplayValue('Test title');
    fireEvent.change(textarea, { target: { value: 'Enter commit' } });
    fireEvent.keyDown(textarea, { key: 'Enter' });

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { title: 'Enter commit' });
  });

  it('clicking body button shows textarea; blur with new value calls onEdit', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar descripción' }));

    const textarea = screen.getByDisplayValue('Test body');
    expect(textarea).toBeInTheDocument();

    fireEvent.change(textarea, { target: { value: 'New body' } });
    fireEvent.blur(textarea);

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { body: 'New body' });
  });

  it('Escape on body textarea reverts without calling onEdit', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Editar descripción' }));

    const textarea = screen.getByDisplayValue('Test body');
    fireEvent.change(textarea, { target: { value: 'Discard body' } });
    fireEvent.keyDown(textarea, { key: 'Escape' });

    expect(screen.getByRole('button', { name: 'Editar descripción' })).toBeInTheDocument();
    expect(handlers.onEdit).not.toHaveBeenCalled();
  });

  it('color button toggles color picker (swatches appear)', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    expect(screen.queryByLabelText('Color default')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Cambiar color' }));

    expect(screen.getByLabelText('Color default')).toBeInTheDocument();
  });

  it('clicking a color swatch calls onEdit with the color and closes picker', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={baseTodo} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Cambiar color' }));
    fireEvent.click(screen.getByLabelText('Color red'));

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { color: 'red' });
    // picker should be closed after selection
    expect(screen.queryByLabelText('Color red')).not.toBeInTheDocument();
  });

  it('done=true card has note-card--done class', () => {
    const handlers = makeHandlers();
    const { container } = render(<NoteCard todo={{ ...baseTodo, done: true }} {...handlers} />);

    // biome-ignore lint/style/noNonNullAssertion: container always has a first child in this test
    expect(container.firstChild!).toHaveClass('note-card--done');
  });

  it('pinned card shows Desfijar nota button label', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={{ ...baseTodo, pinned: true }} {...handlers} />);

    expect(screen.getByRole('button', { name: 'Desfijar nota' })).toBeInTheDocument();
  });

  it('done button calls onEdit with done: false when todo is already done', () => {
    const handlers = makeHandlers();
    render(<NoteCard todo={{ ...baseTodo, done: true }} {...handlers} />);

    fireEvent.click(screen.getByRole('button', { name: 'Marcar pendiente' }));

    expect(handlers.onEdit).toHaveBeenCalledWith(1, { done: false });
  });
});
