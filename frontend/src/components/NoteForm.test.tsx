import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { NoteForm } from './NoteForm';

describe('NoteForm', () => {
  it('renders placeholder "Tomar una nota..."', () => {
    render(<NoteForm onAdd={vi.fn()} />);

    expect(screen.getByText('Tomar una nota...')).toBeInTheDocument();
  });

  it('clicking placeholder shows title input and body textarea', () => {
    render(<NoteForm onAdd={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));

    expect(screen.getByPlaceholderText('Título')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Crear nota...')).toBeInTheDocument();
  });

  it('fill title + body, click Cerrar → onAdd called with correct args', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));

    fireEvent.change(screen.getByPlaceholderText('Título'), { target: { value: 'My title' } });
    fireEvent.change(screen.getByPlaceholderText('Crear nota...'), {
      target: { value: 'My body' },
    });
    fireEvent.click(screen.getByRole('button', { name: /cerrar/i }));

    expect(onAdd).toHaveBeenCalledWith('My title', 'My body', 'default');
  });

  it('fill body only (no title), click Cerrar → onAdd NOT called', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));
    fireEvent.change(screen.getByPlaceholderText('Crear nota...'), {
      target: { value: 'body only' },
    });
    fireEvent.click(screen.getByRole('button', { name: /cerrar/i }));

    expect(onAdd).not.toHaveBeenCalled();
  });

  it('press Escape → form collapses, onAdd NOT called', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));

    const titleInput = screen.getByPlaceholderText('Título');
    fireEvent.change(titleInput, { target: { value: 'Discard me' } });
    fireEvent.keyDown(titleInput, { key: 'Escape' });

    expect(screen.getByText('Tomar una nota...')).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it('color button toggles color picker', () => {
    render(<NoteForm onAdd={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));

    expect(screen.queryByLabelText('Color red')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Cambiar color de la nota' }));

    expect(screen.getByLabelText('Color red')).toBeInTheDocument();
  });

  it('click color swatch → picker closes; subsequent commit uses that color', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cambiar color de la nota' }));
    fireEvent.click(screen.getByLabelText('Color yellow'));

    // Picker should close after selection
    expect(screen.queryByLabelText('Color yellow')).not.toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('Título'), { target: { value: 'Colored note' } });
    fireEvent.click(screen.getByRole('button', { name: /cerrar/i }));

    expect(onAdd).toHaveBeenCalledWith('Colored note', '', 'yellow');
  });

  it('click outside (mousedown on document.body) → commit called, form closes', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));
    fireEvent.change(screen.getByPlaceholderText('Título'), { target: { value: 'Outside note' } });

    fireEvent.mouseDown(document.body);

    expect(onAdd).toHaveBeenCalledWith('Outside note', '', 'default');
    // Form should collapse back to placeholder
    expect(screen.getByText('Tomar una nota...')).toBeInTheDocument();
  });

  it('click outside without title → form closes, onAdd NOT called', () => {
    const onAdd = vi.fn();
    render(<NoteForm onAdd={onAdd} />);

    fireEvent.click(screen.getByRole('button', { name: 'Nueva nota' }));

    fireEvent.mouseDown(document.body);

    expect(onAdd).not.toHaveBeenCalled();
    expect(screen.getByText('Tomar una nota...')).toBeInTheDocument();
  });
});
